package router

import "sync/atomic"

// Store holds the current routing table for lock-free reads.
type Store struct {
	v atomic.Pointer[Table]
}

// Set replaces the active table (may be nil).
func (s *Store) Set(t *Table) {
	s.v.Store(t)
}

// Get returns the active table or nil.
func (s *Store) Get() *Table {
	return s.v.Load()
}
