//go:build windows

package sys

import (
	"fadacontrol/pkg/utils"
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const maxPathSize = 32 * 1024

type windowsSvc struct {
	errSync      sync.Mutex
	handler      SvcHandler
	stopStartErr error
	Name         string
	Option       utils.KeyValue
}
type windowsSystem struct{}

func init() {
	system = windowsSystem{}
}
func (windowsSystem) New(h SvcHandler) (Svc, error) {
	w := &windowsSvc{
		handler: h,
		Name:    ServiceName,
	}

	return w, nil
}

func (w *windowsSvc) Install(execPath string, args ...string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(w.Name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", w.Name)
	}

	startType := mgr.StartAutomatic
	serviceType := windows.SERVICE_WIN32_OWN_PROCESS
	s, err = m.CreateService(w.Name, execPath, mgr.Config{
		DisplayName:      ServiceName,
		Description:      Description,
		StartType:        uint32(startType),
		ServiceStartName: "Localsystem",
		Dependencies:     []string{},
		DelayedAutoStart: false,
		ServiceType:      uint32(serviceType),
	},
		args...,
	)
	if err != nil {
		return err
	}
	err = eventlog.InstallAsEventCreate(w.Name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		if !strings.Contains(err.Error(), "exists") {
			s.Delete()
			return fmt.Errorf("SetupEventLogSource() failed: %s", err)
		}
	}
	return nil
}

func (w *windowsSvc) Uninstall() error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(w.Name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", w.Name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(w.Name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}

func (w *windowsSvc) setError(err error) {
	w.errSync.Lock()
	defer w.errSync.Unlock()
	w.stopStartErr = err
}
func getStopTimeout() time.Duration {
	// For default and paths see https://support.microsoft.com/en-us/kb/146092
	defaultTimeout := time.Millisecond * 20000
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SYSTEM\CurrentControlSet\Control`, registry.READ)
	if err != nil {
		return defaultTimeout
	}
	sv, _, err := key.GetStringValue("WaitToKillServiceTimeout")
	if err != nil {
		return defaultTimeout
	}
	v, err := strconv.Atoi(sv)
	if err != nil {
		return defaultTimeout
	}
	return time.Millisecond * time.Duration(v)
}
func (w *windowsSvc) getError() error {
	w.errSync.Lock()
	defer w.errSync.Unlock()
	return w.stopStartErr
}
func (w *windowsSvc) Start() error {
	m, err := lowPrivMgr()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := lowPrivSvc(m, w.Name)
	if err != nil {
		return err
	}
	defer s.Close()
	return s.Start()
}

func (w *windowsSvc) Stop() error {
	m, err := lowPrivMgr()
	if err != nil {
		return err
	}
	defer m.Disconnect()

	s, err := lowPrivSvc(m, w.Name)
	if err != nil {
		return err
	}
	defer s.Close()

	return w.stopWait(s)
}
func (w *windowsSvc) stopWait(s *mgr.Service) error {

	status, err := s.Control(svc.Stop)
	if err != nil {
		return err
	}

	timeDuration := time.Millisecond * 50

	timeout := time.After(getStopTimeout() + (timeDuration * 2))
	tick := time.NewTicker(timeDuration)
	defer tick.Stop()

	for status.State != svc.Stopped {
		select {
		case <-tick.C:
			status, err = s.Query()
			if err != nil {
				return err
			}
		case <-timeout:
			break
		}
	}
	return nil
}

var interactive = false

func (w *windowsSvc) Run() error {
	w.setError(nil)
	if !interactive {
		// Return error messages from start and stop routines
		// that get executed in the Execute method.
		// Guarded with a mutex as it may run a different thread
		// (callback from windows).
		runErr := svc.Run(w.Name, w)
		startStopErr := w.getError()
		if startStopErr != nil {
			return startStopErr
		}
		if runErr != nil {
			return runErr
		}
		return nil
	}
	err := w.handler.Start(w)
	if err != nil {
		return err
	}

	sigChan := make(chan os.Signal)

	signal.Notify(sigChan, os.Interrupt)

	<-sigChan

	return w.handler.Stop(w)
}

func (w *windowsSvc) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (bool, uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}

	if err := w.handler.Start(w); err != nil {
		w.setError(err)
		return true, 1
	}

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
loop:
	for {
		c := <-r
		switch c.Cmd {
		case svc.Interrogate:
			changes <- c.CurrentStatus
		case svc.Stop:
			changes <- svc.Status{State: svc.StopPending}
			if err := w.handler.Stop(w); err != nil {
				w.setError(err)
				return true, 2
			}
			break loop
		case svc.Shutdown:
			changes <- svc.Status{State: svc.StopPending}
			var err error
			if wsShutdown, ok := w.handler.(Shutdowner); ok {
				err = wsShutdown.Shutdown(w)
			} else {
				err = w.handler.Stop(w)
			}
			if err != nil {
				w.setError(err)
				return true, 2
			}
			break loop
		default:
			continue loop
		}
	}

	return false, 0
}

type WindowsLogger struct {
	ev   *eventlog.Log
	errs chan<- error
}

// Error logs an error message.
func (l WindowsLogger) Error(v ...interface{}) error {
	return l.send(l.ev.Error(3, fmt.Sprint(v...)))
}

// Warning logs an warning message.
func (l WindowsLogger) Warning(v ...interface{}) error {
	return l.send(l.ev.Warning(2, fmt.Sprint(v...)))
}

// Info logs an info message.
func (l WindowsLogger) Info(v ...interface{}) error {
	return l.send(l.ev.Info(1, fmt.Sprint(v...)))
}

// Errorf logs an error message.
func (l WindowsLogger) Errorf(format string, a ...interface{}) error {
	return l.send(l.ev.Error(3, fmt.Sprintf(format, a...)))
}

// Warningf logs an warning message.
func (l WindowsLogger) Warningf(format string, a ...interface{}) error {
	return l.send(l.ev.Warning(2, fmt.Sprintf(format, a...)))
}

// Infof logs an info message.
func (l WindowsLogger) Infof(format string, a ...interface{}) error {
	return l.send(l.ev.Info(1, fmt.Sprintf(format, a...)))
}

// NError logs an error message and an event ID.
func (l WindowsLogger) NError(eventID uint32, v ...interface{}) error {
	return l.send(l.ev.Error(eventID, fmt.Sprint(v...)))
}

// NWarning logs an warning message and an event ID.
func (l WindowsLogger) NWarning(eventID uint32, v ...interface{}) error {
	return l.send(l.ev.Warning(eventID, fmt.Sprint(v...)))
}

// NInfo logs an info message and an event ID.
func (l WindowsLogger) NInfo(eventID uint32, v ...interface{}) error {
	return l.send(l.ev.Info(eventID, fmt.Sprint(v...)))
}

// NErrorf logs an error message and an event ID.
func (l WindowsLogger) NErrorf(eventID uint32, format string, a ...interface{}) error {
	return l.send(l.ev.Error(eventID, fmt.Sprintf(format, a...)))
}

// NWarningf logs an warning message and an event ID.
func (l WindowsLogger) NWarningf(eventID uint32, format string, a ...interface{}) error {
	return l.send(l.ev.Warning(eventID, fmt.Sprintf(format, a...)))
}

// NInfof logs an info message and an event ID.
func (l WindowsLogger) NInfof(eventID uint32, format string, a ...interface{}) error {
	return l.send(l.ev.Info(eventID, fmt.Sprintf(format, a...)))
}
func (l WindowsLogger) send(err error) error {
	if err == nil {
		return nil
	}
	if l.errs != nil {
		l.errs <- err
	}
	return err
}
func lowPrivMgr() (*mgr.Mgr, error) {
	h, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT|windows.SC_MANAGER_ENUMERATE_SERVICE)
	if err != nil {
		return nil, err
	}
	return &mgr.Mgr{Handle: h}, nil
}
func lowPrivSvc(m *mgr.Mgr, name string) (*mgr.Service, error) {
	h, err := windows.OpenService(
		m.Handle, syscall.StringToUTF16Ptr(name),
		windows.SERVICE_QUERY_CONFIG|windows.SERVICE_QUERY_STATUS|windows.SERVICE_START|windows.SERVICE_STOP)
	if err != nil {
		return nil, err
	}
	return &mgr.Service{Handle: h, Name: name}, nil
}
