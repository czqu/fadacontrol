package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

var ErrCapacityOverFlow error

type Cache[K comparable, V any] interface {
	Get(key K) (value V, exist bool)
	Set(key K, value V) error
	SetWithTTL(key K, value V, ttl time.Duration) error
	Delete(key K)
	Size() int64
	Clear()
	Destroy()
	Exists(key K) bool
	StartAutoClean(interval time.Duration)
}
type Store[K comparable, V any] interface {
	Load(key K) (value V, exist bool)
	LoadOrStore(key K, value V) (actual V, loaded bool)
	Store(key K, value V)
	Delete(key K)
	Exists(key K) bool
	Range(f func(key K, value V) bool)
	Size() int64
	Clear()
}
type expireValue[T any] struct {
	Value      T
	Expiration int64
}

func (v *expireValue[T]) Expired(now int64) bool {
	if v.Expiration == -1 {
		return false
	}
	return now >= v.Expiration
}
func (v *expireValue[T]) SetTTL(ttl int64) {
	if ttl < 0 {
		v.Expiration = -1
		return
	}
	v.Expiration = time.Now().UnixNano() + ttl
}

type SyncMap[K comparable, V any] struct {
	m    sync.Map
	size int64
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{}
}

func (s *SyncMap[K, V]) Load(key K) (value V, exist bool) {
	v, ok := s.m.Load(key)
	if !ok {
		var zero V
		return zero, false
	}
	return v.(V), true
}
func (s *SyncMap[K, V]) LoadOrStore(key K, value V) (actual V, loaded bool) {
	actualInterface, loaded := s.m.LoadOrStore(key, value)
	if !loaded {
		atomic.AddInt64(&s.size, 1)
	}
	return actualInterface.(V), loaded
}
func (s *SyncMap[K, V]) Store(key K, value V) {
	if !s.Exists(key) {
		s.m.Store(key, value)
		atomic.AddInt64(&s.size, 1)
	}
	s.m.Store(key, value)
}

func (s *SyncMap[K, V]) Delete(key K) {
	_, ok := s.m.Load(key)
	if ok {
		atomic.AddInt64(&s.size, -1)
	}
	s.m.Delete(key)
}
func (s *SyncMap[K, V]) Exists(key K) bool {
	_, ok := s.m.Load(key)
	return ok
}

func (s *SyncMap[K, V]) Range(f func(key K, value V) bool) {
	s.m.Range(func(key, value interface{}) bool {
		return f(key.(K), value.(V))
	})
}

func (s *SyncMap[K, V]) Size() int64 {
	size := atomic.LoadInt64(&s.size)
	if size < 0 {
		size = 0
	}
	return size
}

func (s *SyncMap[K, V]) Clear() {
	s.m = sync.Map{}
	atomic.StoreInt64(&s.size, 0)
}

type MemCache[K comparable, V any] struct {
	capacity  int64
	store     Store[K, expireValue[V]]
	onceClean sync.Once
	onceClose sync.Once
	exit      chan int
}

func NewSyncMapMemCache[K comparable, V any](capacity int64) *MemCache[K, V] {
	store := NewSyncMap[K, expireValue[V]]()
	return &MemCache[K, V]{
		capacity: capacity,
		store:    store,
		exit:     make(chan int),
	}
}
func (m *MemCache[K, V]) Get(key K) (value V, exist bool) {
	v, ok := m.getValue(key)
	if !ok {
		var zero V
		return zero, false
	}
	return v.Value, true
}
func (m *MemCache[K, V]) Exists(key K) bool {
	_, ok := m.getValue(key)
	return ok
}
func (m *MemCache[K, V]) Set(key K, value V) error {
	return m.SetWithMilliTTL(key, value, -1)
}
func (m *MemCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) error {
	return m.SetWithNanoTTL(key, value, int64(ttl/time.Nanosecond))
}
func (m *MemCache[K, V]) SetWithMicroTTL(key K, value V, ttl int64) error {
	return m.SetWithNanoTTL(key, value, ttl*1000)
}
func (m *MemCache[K, V]) SetWithMilliTTL(key K, value V, ttl int64) error {
	return m.SetWithNanoTTL(key, value, ttl*1000*1000)
}
func (m *MemCache[K, V]) SetWithNanoTTL(key K, value V, ttl int64) error {
	if m.capacity >= 0 {
		if m.Size() >= m.capacity {
			return ErrCapacityOverFlow
		}
	}
	m.setValue(key, value, ttl)
	return nil
}
func (m *MemCache[K, V]) setValue(key K, value V, ttl int64) {
	if ttl == 0 {
		return
	}
	v := expireValue[V]{
		Value: value,
	}
	v.SetTTL(ttl)
	m.store.Store(key, v)
}
func (m *MemCache[K, V]) Delete(key K) {
	m.store.Delete(key)
}
func (m *MemCache[K, V]) Size() int64 {
	return m.store.Size()
}
func (m *MemCache[K, V]) getValue(key K) (value expireValue[V], exist bool) {
	v, ok := m.store.Load(key)
	if !ok {
		return expireValue[V]{}, false
	}

	if v.Expired(time.Now().UnixNano()) {
		return expireValue[V]{}, false
	}
	return v, true
}
func (m *MemCache[K, V]) StartAutoClean(interval time.Duration) {
	m.onceClean.Do(func() {
		go func() {
			ticker := time.NewTicker(interval)
			for {
				select {
				case <-ticker.C:
					m.cleanExpired()
				case <-m.exit:
					ticker.Stop()
					return
				}
			}
		}()
	})
}
func (m *MemCache[K, V]) Destroy() {
	m.onceClose.Do(func() {
		m.store.Clear()
		m.exit <- 1
		close(m.exit)

	})

}
func (m *MemCache[K, V]) Clear() {
	m.store.Clear()
}
func (m *MemCache[K, V]) cleanExpired() {
	now := time.Now().UnixNano()
	m.store.Range(func(k K, v expireValue[V]) bool {
		if v.Expired(now) {
			m.store.Delete(k)
		}
		return true
	})
}
