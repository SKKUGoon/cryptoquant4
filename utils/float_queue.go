package utils

// Simple FIFO queue for float64 values with fixed size
type FloatQueue struct {
	Items []float64
	size  int
}

// NewFloatQueue creates a new FloatQueue with the specified size
func NewFloatQueue(size int) *FloatQueue {
	return &FloatQueue{
		Items: make([]float64, 0, size),
		size:  size,
	}
}

// GetSize returns the size of the queue
func (q *FloatQueue) GetSize() int {
	return q.size
}

// Enqueue adds an item to the end of the queue
func (q *FloatQueue) Enqueue(item float64) {
	if len(q.Items) == q.GetSize() {
		q.Items = q.Items[1:]
	}
	q.Items = append(q.Items, item)
}

// Dequeue removes and returns the item from the front of the queue
func (q *FloatQueue) Dequeue() (float64, bool) {
	if len(q.Items) == 0 {
		return 0, false
	}

	item := q.Items[0]
	q.Items = q.Items[1:]
	return item, true
}

// PeekFirst returns the item at the front of the queue without removing it
func (q *FloatQueue) PeekFirst() (float64, bool) {
	if len(q.Items) == 0 {
		return 0, false
	}

	return q.Items[0], true
}

// PeekLast returns the item at the end of the queue without removing it
func (q *FloatQueue) PeekLast() (float64, bool) {
	if len(q.Items) == 0 {
		return 0, false
	}

	return q.Items[len(q.Items)-1], true
}

// Length returns the number of items in the queue
func (q *FloatQueue) Length() int {
	return len(q.Items)
}

// IsFull returns true if the queue has reached its size limit
func (q *FloatQueue) IsFull() bool {
	return len(q.Items) == q.GetSize()
}
