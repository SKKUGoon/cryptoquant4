package utils_test

import (
	"testing"

	"cryptoquant.com/m/utils"
)

func TestNewFloatQueue(t *testing.T) {
	size := 5
	q := utils.NewFloatQueue(size)

	if q.GetSize() != size {
		t.Errorf("Expected size %d, got %d", size, q.GetSize())
	}

	if len(q.Items) != 0 {
		t.Errorf("Expected empty queue, got length %d", len(q.Items))
	}

	if cap(q.Items) != size {
		t.Errorf("Expected capacity %d, got %d", size, cap(q.Items))
	}
}

func TestEnqueue(t *testing.T) {
	q := utils.NewFloatQueue(3)

	// Test basic enqueue
	q.Enqueue(1.0)
	if len(q.Items) != 1 || q.Items[0] != 1.0 {
		t.Errorf("Enqueue failed, got %v", q.Items)
	}

	// Test multiple enqueues
	q.Enqueue(2.0)
	q.Enqueue(3.0)
	if len(q.Items) != 3 {
		t.Errorf("Expected length 3, got %d", len(q.Items))
	}

	// Test overflow
	q.Enqueue(4.0)
	if len(q.Items) != 3 {
		t.Errorf("Expected length 3 after overflow, got %d", len(q.Items))
	}
	if q.Items[0] != 2.0 || q.Items[2] != 4.0 {
		t.Errorf("Unexpected items after overflow: %v", q.Items)
	}
}

func TestDequeue(t *testing.T) {
	q := utils.NewFloatQueue(3)

	// Test empty queue
	val, ok := q.Dequeue()
	if ok || val != 0 {
		t.Errorf("Expected (0, false) from empty queue, got (%f, %v)", val, ok)
	}

	// Test with items
	q.Enqueue(1.0)
	q.Enqueue(2.0)

	val, ok = q.Dequeue()
	if !ok || val != 1.0 {
		t.Errorf("Expected (1.0, true), got (%f, %v)", val, ok)
	}
	if len(q.Items) != 1 {
		t.Errorf("Expected length 1 after dequeue, got %d", len(q.Items))
	}
}

func TestPeekFirst(t *testing.T) {
	q := utils.NewFloatQueue(3)

	// Test empty queue
	val, ok := q.PeekFirst()
	if ok || val != 0 {
		t.Errorf("Expected (0, false) from empty queue, got (%f, %v)", val, ok)
	}

	// Test with items
	q.Enqueue(1.0)
	q.Enqueue(2.0)

	val, ok = q.PeekFirst()
	if !ok || val != 1.0 {
		t.Errorf("Expected (1.0, true), got (%f, %v)", val, ok)
	}
	if len(q.Items) != 2 {
		t.Errorf("Expected length unchanged after peek, got %d", len(q.Items))
	}
}

func TestPeekLast(t *testing.T) {
	q := utils.NewFloatQueue(3)

	// Test empty queue
	val, ok := q.PeekLast()
	if ok || val != 0 {
		t.Errorf("Expected (0, false) from empty queue, got (%f, %v)", val, ok)
	}

	// Test with items
	q.Enqueue(1.0)
	q.Enqueue(2.0)

	val, ok = q.PeekLast()
	if !ok || val != 2.0 {
		t.Errorf("Expected (2.0, true), got (%f, %v)", val, ok)
	}
	if len(q.Items) != 2 {
		t.Errorf("Expected length unchanged after peek, got %d", len(q.Items))
	}
}

func TestLength(t *testing.T) {
	q := utils.NewFloatQueue(3)

	if q.Length() != 0 {
		t.Errorf("Expected length 0, got %d", q.Length())
	}

	q.Enqueue(1.0)
	if q.Length() != 1 {
		t.Errorf("Expected length 1, got %d", q.Length())
	}

	q.Enqueue(2.0)
	q.Enqueue(3.0)
	if q.Length() != 3 {
		t.Errorf("Expected length 3, got %d", q.Length())
	}

	q.Enqueue(4.0) // Should maintain size
	if q.Length() != 3 {
		t.Errorf("Expected length 3 after overflow, got %d", q.Length())
	}
}

func TestIsFull(t *testing.T) {
	q := utils.NewFloatQueue(3)

	if q.IsFull() {
		t.Error("Expected queue not to be full")
	}

	q.Enqueue(1.0)
	q.Enqueue(2.0)
	if q.IsFull() {
		t.Error("Expected queue not to be full")
	}

	q.Enqueue(3.0)
	if !q.IsFull() {
		t.Error("Expected queue to be full")
	}

	q.Enqueue(4.0) // Should maintain full state
	if !q.IsFull() {
		t.Error("Expected queue to remain full after overflow")
	}
}
