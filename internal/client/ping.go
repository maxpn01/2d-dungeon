package client

import (
	"context"
	"time"

	"github.com/coder/websocket"
)

func (s *Session) Ping() (time.Duration, bool) {
	s.pingMu.RLock()
	defer s.pingMu.RUnlock()
	return s.ping, !s.pingAt.IsZero() && time.Since(s.pingAt) < 5*time.Second
}

func (s *Session) measurePing(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		started := time.Now()
		err := conn.Ping(pingCtx)
		elapsed := time.Since(started)
		cancel()
		if err == nil {
			s.pingMu.Lock()
			s.ping, s.pingAt = elapsed, time.Now()
			s.pingMu.Unlock()
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
