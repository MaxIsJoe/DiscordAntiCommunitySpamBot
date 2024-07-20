package main

import (
	"time"
)

type ErrorEntry struct {
	Message   string
	Timestamp time.Time
}

type ErrorBuffer struct {
	errors []ErrorEntry
	size   int
	start  int
	count  int
}

func NewErrorBuffer(size int) *ErrorBuffer {
	return &ErrorBuffer{
		errors: make([]ErrorEntry, size),
		size:   size,
	}
}

func (b *ErrorBuffer) AddError(message string) {
	entry := ErrorEntry{
		Message:   message,
		Timestamp: time.Now(),
	}
	if b.count < b.size {
		b.errors[(b.start+b.count)%b.size] = entry
		b.count++
	} else {
		b.errors[b.start] = entry
		b.start = (b.start + 1) % b.size
	}
}

func (b *ErrorBuffer) GetErrors() []ErrorEntry {
	result := make([]ErrorEntry, b.count)
	for i := 0; i < b.count; i++ {
		result[i] = b.errors[(b.start+i)%b.size]
	}
	return result
}
