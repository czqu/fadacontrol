package utils

import "testing"

func testRingBuffer(t *testing.T, buffer RingBuffer) {

	buffer.Write(1)
	buffer.Write(2)
	buffer.Write(3)

	if buffer.Full() != true {
		t.Errorf("Expected buffer to be full")
	}

	if val := buffer.ReadAndRemove(); val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}

	buffer.Write(4)

	if val := buffer.ReadAndRemove(); val != 2 {
		t.Errorf("Expected 2, got %v", val)
	}

	buffer.Write(5)
	buffer.Write(6)
	all := buffer.ReadAll()
	expected := []any{4, 5, 6, 4}
	for i, v := range all {
		if v != expected[i] {
			t.Errorf("Expected %v, got %v", expected[i], v)
		}
	}

	if buffer.Full() != true {
		t.Errorf("Expected buffer to be not full")
	}

}
func testRingBufferEdgeCases(t *testing.T, buffer RingBuffer) {

	if val := buffer.ReadAndRemove(); val != nil {
		t.Errorf("Expected nil, got %v", val)
	}

	all := buffer.ReadAll()
	if len(all) != 0 {
		t.Errorf("Expected empty slice, got %v", all)
	}

	buffer.Write(1)
	if val := buffer.ReadAndRemove(); val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}

	buffer.Write(1)
	buffer.Write(2)
	buffer.Write(3)
	buffer.Write(4)
	if val := buffer.ReadAndRemove(); val != 2 {
		t.Errorf("Expected 2, got %v", val)
	}
}
func testMixedOperations(t *testing.T, buffer RingBuffer) {
	buffer.Write(1)
	buffer.Write(2)
	if val := buffer.ReadAndRemove(); val != 1 {
		t.Errorf("Expected 1, got %v", val)
	}
	buffer.Write(3)
	if val := buffer.ReadAndRemove(); val != 2 {
		t.Errorf("Expected 2, got %v", val)
	}
	buffer.Write(4)
	buffer.Write(5)
	if val := buffer.ReadAndRemove(); val != 3 {
		t.Errorf("Expected 3, got %v", val)
	}
	if val := buffer.ReadAndRemove(); val != 4 {
		t.Errorf("Expected 4, got %v", val)
	}
	if val := buffer.ReadAndRemove(); val != 5 {
		t.Errorf("Expected 5, got %v", val)
	}
	buffer.ReadAll()
	if val := buffer.ReadAndRemove(); val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
}
func testWriteAndReadAll(t *testing.T, buffer RingBuffer) {
	buffer.Write(1)
	buffer.Write(2)
	buffer.Write(3)
	all := buffer.ReadAll()
	expected := []any{1, 2, 3}
	for i, v := range all {
		if v != expected[i] {
			t.Errorf("Expected %v, got %v", expected[i], v)
		}
	}
}
func testOverwriteAndReadAll(t *testing.T, buffer RingBuffer) {
	buffer.Write(1)
	buffer.Write(2)
	buffer.Write(3)
	buffer.Write(4)
	all := buffer.ReadAll()
	expected := []any{2, 3, 4}
	for i, v := range all {
		if v != expected[i] {
			t.Errorf("Expected %v, got %v", expected[i], v)
		}
	}
}
func TestLinkedListRingBuffer(t *testing.T) {
	buffer := NewLinkedListRingBuffer(3)
	testRingBuffer(t, buffer)
	buffer = NewLinkedListRingBuffer(3)
	testRingBufferEdgeCases(t, buffer)
	buffer = NewLinkedListRingBuffer(3)
	testMixedOperations(t, buffer)
	buffer = NewLinkedListRingBuffer(3)
	testWriteAndReadAll(t, buffer)
	buffer = NewLinkedListRingBuffer(3)
	testOverwriteAndReadAll(t, buffer)
}

func TestArrayRingBuffer(t *testing.T) {
	buffer := NewArrayRingBuffer(3)
	testRingBuffer(t, buffer)
	buffer = NewArrayRingBuffer(3)
	testRingBufferEdgeCases(t, buffer)
	buffer = NewArrayRingBuffer(3)
	testMixedOperations(t, buffer)
	buffer = NewArrayRingBuffer(3)
	testWriteAndReadAll(t, buffer)
	buffer = NewArrayRingBuffer(3)
	testOverwriteAndReadAll(t, buffer)
}
func BenchmarkLinkedListRingBufferWrite(b *testing.B) {
	buffer := NewLinkedListRingBuffer(10000000)
	for i := 0; i < b.N; i++ {
		buffer.Write(i)
	}
}

func BenchmarkLinkedListRingBufferRead(b *testing.B) {
	buffer := NewLinkedListRingBuffer(10000000)
	for i := 0; i < 10000000; i++ {
		buffer.Write(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.ReadAndRemove()
	}
}
func BenchmarkLinkedListRingBuffer_ReadAll(b *testing.B) {
	buffer := NewLinkedListRingBuffer(10000000)
	for i := 0; i < 10000000; i++ {
		buffer.Write(i)
	}
	b.ResetTimer()
	buffer.ReadAll()
}
func BenchmarkArrayRingBufferWrite(b *testing.B) {
	buffer := NewArrayRingBuffer(10000000)
	for i := 0; i < b.N; i++ {
		buffer.Write(i)
	}
}

func BenchmarkArrayRingBufferRead(b *testing.B) {
	buffer := NewArrayRingBuffer(10000000)
	for i := 0; i < 10000000; i++ {
		buffer.Write(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.ReadAndRemove()
	}
}
func BenchmarkArrayRingBuffer_ReadAll(b *testing.B) {
	buffer := NewArrayRingBuffer(10000000)
	for i := 0; i < 10000000; i++ {
		buffer.Write(i)
	}
	b.ResetTimer()
	buffer.ReadAll()
}
func BenchmarkArrayRingBufferReadSmall(b *testing.B) {
	buffer := NewArrayRingBuffer(1000)
	for i := 0; i < 10000000; i++ {
		buffer.Write(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.ReadAndRemove()
	}
}
func BenchmarkArrayRingBuffer_ReadAllSmall(b *testing.B) {
	buffer := NewArrayRingBuffer(1000)
	for i := 0; i < 10000000; i++ {
		buffer.Write(i)
	}
	b.ResetTimer()
	buffer.ReadAll()
}
func BenchmarkLinkedListRingBufferReadSmall(b *testing.B) {
	buffer := NewLinkedListRingBuffer(1000)
	for i := 0; i < 10000000; i++ {
		buffer.Write(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buffer.ReadAndRemove()
	}
}
func BenchmarkLinkedListRingBuffer_ReadAllSmall(b *testing.B) {
	buffer := NewLinkedListRingBuffer(1000)
	for i := 0; i < 10000000; i++ {
		buffer.Write(i)
	}
	b.ResetTimer()
	buffer.ReadAll()
}
