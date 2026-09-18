package gui

import (
	"sync"
	"time"
)

type Stats struct {
	BytesIn   uint64
	BytesOut  uint64
	SpeedIn   float64
	SpeedOut  float64
	StartTime time.Time
	mu        sync.RWMutex
	lastIn    uint64
	lastOut   uint64
	lastTick  time.Time
}

func NewStats() *Stats {
	return &Stats{
		StartTime: time.Now(),
		lastTick:  time.Now(),
	}
}

func (s *Stats) AddIn(n uint64) {
	s.mu.Lock()
	s.BytesIn += n
	s.mu.Unlock()
}

func (s *Stats) AddOut(n uint64) {
	s.mu.Lock()
	s.BytesOut += n
	s.mu.Unlock()
}

func (s *Stats) Tick() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(s.lastTick).Seconds()
	if elapsed < 0.5 {
		return
	}

	s.SpeedIn = float64(s.BytesIn-s.lastIn) / elapsed
	s.SpeedOut = float64(s.BytesOut-s.lastOut) / elapsed
	s.lastIn = s.BytesIn
	s.lastOut = s.BytesOut
	s.lastTick = now
}

func (s *Stats) Snapshot() (in, out uint64, speedIn, speedOut float64, uptime time.Duration) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.BytesIn, s.BytesOut, s.SpeedIn, s.SpeedOut, time.Since(s.StartTime)
}
