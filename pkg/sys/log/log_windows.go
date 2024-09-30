package log

import (
	"fadacontrol/internal/base/constants"
	"fmt"
	"golang.org/x/sys/windows/svc/eventlog"
	"time"
)

type SysWindowsLog struct {
	_log  *eventlog.Log
	_skip int
}

func NewLogSysWindows() *SysWindowsLog {
	_log, err := eventlog.Open(constants.ServiceName)
	if err != nil {
		return nil
	}
	return &SysWindowsLog{_log: _log, _skip: 5}
}
func (l *SysWindowsLog) Fatalf(error error, format string, v ...interface{}) {
	l._log.Error(1004, fmt.Sprintf(l.GetLogHead("Fatal")+format, v...))
}
func (l *SysWindowsLog) Errorf(error error, format string, v ...interface{}) {
	l._log.Error(1003, fmt.Sprintf(l.GetLogHead("Errof")+format, v...))
}
func (l *SysWindowsLog) Warnf(error error, format string, v ...interface{}) {
	l._log.Warning(1002, fmt.Sprintf(l.GetLogHead("Warn")+format, v...))
}
func (l *SysWindowsLog) Infof(error error, format string, v ...interface{}) {
	l._log.Info(1001, fmt.Sprintf(l.GetLogHead("Info")+format, v...))
}
func (l *SysWindowsLog) Debugf(error error, format string, v ...interface{}) {
	l._log.Info(1000, fmt.Sprintf(l.GetLogHead("Fatal")+format, v...))
}
func (l *SysWindowsLog) Close() error {
	return l._log.Close()
}
func (l *SysWindowsLog) SetSkip(skip int) {
	l._skip = skip
}
func (l *SysWindowsLog) GetLogHead(loglevel string) string {
	now := time.Now()
	file, line := GetFileLine(l._skip)
	return fmt.Sprintf("[%s] [%s]  %s:%d ", now.Format("2006-01-02 15:04:05"), loglevel, file, line)
}
func init() {
	log = NewLogSysWindows()
}
