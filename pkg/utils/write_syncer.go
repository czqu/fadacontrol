package utils

import (
	"bufio"
	"container/ring"
	"io"
	"net/http"
	"sync"
	"time"
)

type WriteSyncer interface {
	io.Writer
	Sync() error
}

func AddSync(w io.Writer) WriteSyncer {
	switch w := w.(type) {
	case WriteSyncer:
		return w
	default:
		return writerWrapper{w}
	}
}
func AddLockSync(ws WriteSyncer) WriteSyncer {
	if _, ok := ws.(*lockedWriteSyncer); ok {
		// no need to layer on another lock
		return ws
	}
	return &lockedWriteSyncer{ws: ws}
}
func AddResponseSyncer(w http.ResponseWriter) WriteSyncer {
	return &responseWriterSyncer{w: w}
}

type responseWriterSyncer struct {
	w http.ResponseWriter
}

func (r *responseWriterSyncer) Sync() error {
	if r.w == nil {
		return nil
	}
	switch r.w.(type) {
	case http.Flusher:
		if flusher, ok := r.w.(http.Flusher); ok && flusher != nil {
			flusher.Flush()
		}
	default:
		return nil
	}

	return nil
}
func (r *responseWriterSyncer) Write(p []byte) (int, error) {
	if r.w == nil {
		return 0, nil
	}
	return r.w.Write(p)
}

type lockedWriteSyncer struct {
	sync.Mutex
	ws WriteSyncer
}

func (s *lockedWriteSyncer) Sync() error {
	s.Lock()
	err := s.ws.Sync()
	s.Unlock()
	return err
}

func (s *lockedWriteSyncer) Write(bs []byte) (int, error) {
	s.Lock()
	n, err := s.ws.Write(bs)
	s.Unlock()
	return n, err
}

type writerWrapper struct {
	io.Writer
}

func (w writerWrapper) Sync() error {
	return nil
}

type RingBuffer interface {
	Read() any
	Write(any)
	ReadAll() []any
	Full() bool
}
type LinkedListRingBuffer struct {
	buf      *ring.Ring
	readPos  *ring.Ring
	writePos *ring.Ring

	length   int
	capacity int
}

func NewLinkedListRingBuffer(size int) *LinkedListRingBuffer {
	buf := ring.New(size)
	return &LinkedListRingBuffer{buf: buf, capacity: size, length: 0, readPos: buf, writePos: buf}
}
func (l *LinkedListRingBuffer) Read() any {
	if l.length == 0 {
		return nil
	}
	val := l.readPos.Value

	l.readPos = l.readPos.Next()
	l.length = l.length - 1
	return val
}
func (l *LinkedListRingBuffer) ReadAll() []any {
	ret := make([]any, l.length)
	for i := 0; i < l.length; i++ {
		val := l.readPos.Value
		ret[i] = val
		l.readPos = l.readPos.Next()

	}
	l.length = 0

	return ret
}
func (l *LinkedListRingBuffer) Write(val any) {
	l.writePos.Value = val
	l.writePos = l.writePos.Next()

	if l.length == l.capacity {
		l.readPos = l.readPos.Next()

	} else {
		l.length = l.length + 1
	}

}
func (l *LinkedListRingBuffer) Full() bool {
	return l.length == l.capacity
}

type ArrayRingBuffer struct {
	buf      []any
	readPos  int
	writePos int
	length   int
	capacity int
}

func NewArrayRingBuffer(size int) *ArrayRingBuffer {
	return &ArrayRingBuffer{buf: make([]any, size), capacity: size}
}

func (a *ArrayRingBuffer) Read() any {
	if a.length == 0 {
		return nil
	}
	val := a.buf[a.readPos]
	if a.readPos == a.capacity-1 {
		a.readPos = 0
	} else {
		a.readPos = a.readPos + 1
	}
	a.length--
	return val
}

func (a *ArrayRingBuffer) ReadAll() []any {
	ret := make([]any, a.length)
	for i := 0; i < a.length; i++ {
		val := a.buf[a.readPos]
		ret[i] = val
		if a.readPos == a.capacity-1 {
			a.readPos = 0
		} else {
			a.readPos = a.readPos + 1
		}

	}
	a.length = 0
	return ret
}

func (a *ArrayRingBuffer) Write(val any) {
	a.buf[a.writePos] = val
	if a.writePos == a.capacity-1 {
		a.writePos = 0
	} else {
		a.writePos = a.writePos + 1
	}

	if a.length == a.capacity {
		if a.readPos == a.capacity-1 {
			a.readPos = 0
		} else {
			a.readPos = a.readPos + 1
		}
	} else {
		a.length++
	}
}
func (a *ArrayRingBuffer) empty() bool {
	return a.readPos == a.writePos
}
func (a *ArrayRingBuffer) Full() bool {
	return a.length == a.capacity
}

type MultiBufferSyncWriteSyncer struct {
	m    map[int]WriteSyncer
	lock sync.Mutex
	buf  RingBuffer
	size int
}

func NewMultiBufferSyncWriteSyncer(cacheSize int) *MultiBufferSyncWriteSyncer {
	buffer := NewArrayRingBuffer(cacheSize)
	return &MultiBufferSyncWriteSyncer{m: make(map[int]WriteSyncer), lock: sync.Mutex{}, size: cacheSize, buf: buffer}
}
func (mu *MultiBufferSyncWriteSyncer) AddSyncersAndFlushBuf(ws ...WriteSyncer) []int {
	mu.lock.Lock()
	defer mu.lock.Unlock()
	var ids []int
	for _, w := range ws {
		id := mu.AddSyncer(w)
		ids = append(ids, id)
	}

	cache := mu.buf.ReadAll()

	for _, v := range cache {
		value := v.([]byte)
		for _, id := range ids {
			syncer := mu.m[id]
			syncer.Write(value)
			syncer.Sync()
		}
	}
	return ids
}
func (mu *MultiBufferSyncWriteSyncer) AddSyncerAndFlushBuf(ws WriteSyncer) int {
	mu.lock.Lock()
	defer mu.lock.Unlock()
	id := mu.AddSyncer(ws)
	cache := mu.buf.ReadAll()
	for _, v := range cache {
		value := v.([]byte)

		syncer := mu.m[id]
		syncer.Write(value)
		syncer.Sync()

	}

	return id
}
func (mu *MultiBufferSyncWriteSyncer) AddSyncerLock(ws WriteSyncer) int {
	mu.lock.Lock()
	defer mu.lock.Unlock()
	return mu.AddSyncer(ws)

}
func (mu *MultiBufferSyncWriteSyncer) AddSyncer(ws WriteSyncer) int {

	id := len(mu.m)
	mu.m[id] = ws
	return id

}
func (mu *MultiBufferSyncWriteSyncer) Remove(id int) {
	mu.lock.Lock()
	defer mu.lock.Unlock()
	delete(mu.m, id)
}

// See https://golang.org/src/io/multi.go
// will return the smallest the number of bytes been written
func (mu *MultiBufferSyncWriteSyncer) Write(p []byte) (int, error) {
	mu.lock.Lock()
	defer mu.lock.Unlock()
	val := make([]byte, len(p))
	copy(val, p)
	mu.buf.Write(val)
	var writeErr error
	nWritten := 0
	for _, w := range mu.m {
		n, err := w.Write(p)
		writeErr = err
		if nWritten == 0 && n != 0 {
			nWritten = n
		} else if n < nWritten {
			nWritten = n
		}
	}
	return nWritten, writeErr
}

func (mu *MultiBufferSyncWriteSyncer) Sync() error {
	mu.lock.Lock()
	defer mu.lock.Unlock()
	var err error
	for _, w := range mu.m {
		err = w.Sync()
	}
	return err
}

type Clock interface {
	Now() time.Time

	NewTicker(time.Duration) *time.Ticker
}
type systemClock struct{}

func (systemClock) Now() time.Time {
	return time.Now()
}

func (systemClock) NewTicker(duration time.Duration) *time.Ticker {
	return time.NewTicker(duration)
}

type BufferedWriteSyncer struct {
	WS WriteSyncer

	Size int

	FlushInterval time.Duration

	Clock Clock

	mu          sync.Mutex
	initialized bool
	stopped     bool
	writer      *bufio.Writer
	ticker      *time.Ticker
	stop        chan struct{}
	done        chan struct{}
}

const (
	// _defaultBufferSize specifies the default size used by Buffer.
	_defaultBufferSize = 256 * 1024 // 256 kB

	// _defaultFlushInterval specifies the default flush interval for
	// Buffer.
	_defaultFlushInterval = 30 * time.Second
)

func (s *BufferedWriteSyncer) initialize() {
	size := s.Size
	if size == 0 {
		size = _defaultBufferSize
	}

	flushInterval := s.FlushInterval
	if flushInterval == 0 {
		flushInterval = _defaultFlushInterval
	}

	if s.Clock == nil {
		s.Clock = systemClock{}
	}

	s.ticker = s.Clock.NewTicker(flushInterval)
	s.writer = bufio.NewWriterSize(s.WS, size)
	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	s.initialized = true
	go s.flushLoop()
}
func (s *BufferedWriteSyncer) Write(bs []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		s.initialize()
	}

	// To avoid partial writes from being flushed, we manually flush the existing buffer if:
	// * The current write doesn't fit into the buffer fully, and
	// * The buffer is not empty (since bufio will not split large writes when the buffer is empty)
	if len(bs) > s.writer.Available() && s.writer.Buffered() > 0 {
		if err := s.writer.Flush(); err != nil {
			return 0, err
		}
	}

	return s.writer.Write(bs)
}

// Sync flushes buffered log data directly.
func (s *BufferedWriteSyncer) Sync() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var err error
	if s.initialized {
		err = s.writer.Flush()
	}
	err = s.WS.Sync()
	return err
}

// flushLoop flushes the buffer at the configured interval until Stop is
// called.
func (s *BufferedWriteSyncer) flushLoop() {

	defer close(s.done)

	for {
		select {
		case <-s.ticker.C:
			// we just simply ignore error here
			// because the underlying bufio writer stores any errors
			// and we return any error from Sync() as part of the close
			_ = s.Sync()
		case <-s.stop:
			return
		}
	}
}

// Stop closes the buffer, cleans up background goroutines, and flushes
// remaining unwritten data.
func (s *BufferedWriteSyncer) Stop() (err error) {
	var stopped bool

	// Critical section.
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		if !s.initialized {
			return
		}

		stopped = s.stopped
		if stopped {
			return
		}
		s.stopped = true

		s.ticker.Stop()
		close(s.stop) // tell flushLoop to stop
		<-s.done      // and wait until it has
	}()

	// Don't call Sync on consecutive Stops.
	if !stopped {
		err = s.Sync()
	}

	return err
}

type ResponseWriterWrapper struct {
	http.ResponseWriter
	closed bool
}
