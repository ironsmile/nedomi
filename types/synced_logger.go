package types

import "sync"

// SyncLogger is a container for a logger that can be setted and getted
// safely from multiple goroutines
type SyncLogger struct {
	logger Logger
	mut    sync.RWMutex
}

// GetLogger get the Logger
func (s *SyncLogger) GetLogger() (l Logger) {
	s.mut.RLock()
	l = s.logger
	s.mut.RUnlock()
	return l
}

// SetLogger set the Logger
func (s *SyncLogger) SetLogger(l Logger) {
	s.mut.Lock()
	s.logger = l
	s.mut.Unlock()
}
