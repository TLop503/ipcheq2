package router

type ResultsBuffer struct {
	data  []FrontEndData
	head  int // index of newest element
	count int // number of valid elements
}

// NewResultsBuffer creates a fixed-size buffer
func NewResultsBuffer(size int) *ResultsBuffer {
	return &ResultsBuffer{
		data: make([]FrontEndData, size),
		head: -1, // no elements yet
	}
}

// Add inserts a new item as the "newest"
func (r *ResultsBuffer) Add(f FrontEndData) {
	if len(r.data) == 0 {
		return
	}

	// move head forward (wrap around)
	r.head = (r.head + 1) % len(r.data)
	r.data[r.head] = f

	if r.count < len(r.data) {
		r.count++
	}
}

func (r *ResultsBuffer) Slice() []FrontEndData {
	if r.count == 0 {
		return nil
	}

	out := make([]FrontEndData, r.count)

	for i := 0; i < r.count; i++ {
		// walk backwards from head
		idx := (r.head - i + len(r.data)) % len(r.data)
		out[i] = r.data[idx]
	}

	return out
}
