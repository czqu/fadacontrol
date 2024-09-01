package custom_command_schema

import "sync"

type Command struct {
	Name    string            `yaml:"name"`
	Cmd     string            `yaml:"cmd"`
	Args    []string          `yaml:"args"`
	Env     map[string]string `yaml:"env"`
	WorkDir string            `yaml:"workdir"`
}
type CustomWriter struct {
	Ch   chan []byte
	done chan struct{}
	mu   sync.Mutex
}

func NewCustomWriter() *CustomWriter {
	return &CustomWriter{
		Ch:   make(chan []byte),
		done: make(chan struct{}),
	}
}

func (cw *CustomWriter) Write(p []byte) (n int, err error) {
	cw.mu.Lock()
	defer cw.mu.Unlock()

	data := make([]byte, len(p))
	copy(data, p)
	cw.Ch <- data
	return len(p), nil
}
func (cw *CustomWriter) Read() ([]byte, bool) {
	select {
	case data := <-cw.Ch:
		return data, true
	case <-cw.done:
		return nil, false
	}
}

func (cw *CustomWriter) Close() {
	close(cw.done)
}
