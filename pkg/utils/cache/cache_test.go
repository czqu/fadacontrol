package cache

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestMemCache(t *testing.T) {

	cache := NewSyncMapMemCache[string, string](10)
	cache.StartAutoClean(time.Millisecond * 100)

	// Test Set and Get
	if err := cache.SetWithTTL("key1", "value1", 5*time.Second); err != nil {
		t.Fatalf("Expected no error on SetWithTTL, got %v", err)
	}
	if value, ok := cache.Get("key1"); !ok || value != "value1" {
		t.Fatalf("Expected value 'value1', got %v", value)
	}

	// Test TTL Expiration
	cache.SetWithTTL("key2", "value2", 1*time.Second)
	time.Sleep(2 * time.Second)
	if _, ok := cache.Get("key2"); ok {
		t.Fatal("Expected key2 to be expired")
	}

	// Test Size
	cache.SetWithTTL("key3", "value3", 10*time.Second)
	if cache.Size() != 2 {
		t.Fatalf("Expected size 2, got %d", cache.Size())
	}
	cache.Destroy()
	// Test capacity limit
	cache = NewSyncMapMemCache[string, string](1)
	cache.StartAutoClean(time.Millisecond * 100)
	if err := cache.SetWithTTL("key4", "value4", 10*time.Second); err != nil {
		t.Fatalf("Expected no error on SetWithTTL, got %v", err)
	}
	if err := cache.SetWithTTL("key5", "value5", 10*time.Second); err != nil {
		t.Fatalf("Expected no error on SetWithTTL, got %v", err)
	}

	// Test Clear and Destroy

	cache.Destroy()
	if _, ok := cache.Get("key1"); ok {
		t.Fatal("Expected no value after Destroy")
	}

	cache = NewSyncMapMemCache[string, string](10)
	cache.StartAutoClean(time.Second * 100)
	// Test TTL Expiration
	cache.SetWithTTL("key2", "value2", 1*time.Second)
	cache.Set("key3", "value3")
	cache.SetWithTTL("key4", "value4", 0)
	cache.SetWithMicroTTL("key5", "value5", int64((5*time.Second)/time.Microsecond))
	time.Sleep(2 * time.Second)

	if _, ok := cache.Get("key2"); ok {
		t.Fatal("Expected key2 to be expired")
	}
	v, ok := cache.Get("key3")
	if !ok || v != "value3" {
		t.Fatalf("Expected value 'value3', got %v", v)
	}
	cache.Set("key3", "key3_")
	v, ok = cache.Get("key3")
	if !ok || v != "key3_" {
		t.Fatalf("Expected value 'key3_', got %v", v)
	}
	if _, ok := cache.Get("key4"); ok {
		t.Fatal("Expected key4 to be expired")
	}

}

func TestMemCacheConcurrency(t *testing.T) {

	cache := NewSyncMapMemCache[string, int](100)

	var wg sync.WaitGroup
	numRoutines := 100
	numOps := 100

	// Concurrent Set
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				cache.SetWithTTL("key"+strconv.Itoa(i), i+j, 5*time.Second)
			}
		}(i)
	}
	wg.Wait()

	// Concurrent Get and Check existence
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				_, _ = cache.Get("key" + strconv.Itoa(i))
			}
		}(i)
	}
	wg.Wait()

	// Validate size should be within expected limits
	if cache.Size() > int64(numRoutines) {
		t.Fatalf("Expected size <= %d, got %d", numRoutines, cache.Size())
	}

	// Concurrent Delete
	wg.Add(numRoutines)
	for i := 0; i < numRoutines; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				cache.Delete("key" + strconv.Itoa(i))
			}
		}(i)
	}
	wg.Wait()

	if cache.Size() != 0 {
		t.Fatalf("Expected cache size 0 after concurrent deletes, got %d", cache.Size())
	}
}
func TestSyncMap(t *testing.T) {
	syncMap := NewSyncMap[string, string]()

	syncMap.Store("store", "1")
	if !syncMap.Exists("store") {
		t.Fatalf("Expected store to exist")
		return
	}

	v, ok := syncMap.Load("store")
	if !ok {
		t.Fatal("store key not exists")
		return
	}
	if v != "1" {
		t.Fatal("store value not equal")
		return
	}

	ac, exists := syncMap.LoadOrStore("store", "2")
	if !exists {
		t.Fatal("store should exists")
		return
	}

	if ac != "1" {
		t.Fatal("store value should = 1")
		return
	}

	ac, exists = syncMap.LoadOrStore("restore", "2")
	if exists {
		t.Fatal("store should not exists")
		return
	}

	if ac != "2" {
		t.Fatal("store value should = 2")
		return
	}

	syncMap.Range(func(k string, v string) bool {
		switch k {
		case "store", "restore":
		default:
			t.Fatal("key error ", k)
			return false
		}

		return true
	})
	count := 0
	syncMap.Range(func(k string, v string) bool {
		count++
		return false
	})

	if count != 1 {
		t.Fatal("range break error")
	}

	syncMap.Delete("store")

	_, ok = syncMap.Load("store")
	if ok {
		t.Fatal("store key should not exists")
		return
	}
}
