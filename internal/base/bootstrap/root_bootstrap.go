package bootstrap

type Bootstrap interface {
	Start() error
	Stop() error
}
