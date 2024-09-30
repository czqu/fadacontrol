package sys

import "fadacontrol/internal/base/constants"

const (
	ServiceName = constants.ServiceName
	Description = "FadaControl Service"
)

type Svc interface {
	Start() error
	Run() error
	Stop() error
	Install(args ...string) error
	Uninstall() error
}
type SvcHandler interface {
	Start(s Svc) error

	Stop(s Svc) error
}
type System interface {
	// New creates a new service for this system.
	New(h SvcHandler) (Svc, error)
}

// Shutdowner represents a service interface for a program that differentiates between "stop" and
// "shutdown". A shutdown is triggered when the whole box (not just the service) is stopped.
type Shutdowner interface {
	SvcHandler
	// Shutdown provides a place to clean up program execution when the system is being shutdown.
	// It is essentially the same as Stop but for the case where machine is being shutdown/restarted
	// instead of just normally stopping the service. Stop won't be called when Shutdown is.
	Shutdown(s Svc) error
}

var (
	system System
)

func New(i SvcHandler) (Svc, error) {
	return system.New(i)
}
