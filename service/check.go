package service

import (
	"context"
	"time"
)

func (s *service) startCheck() {
	s.checkMtx.Lock()
	defer s.checkMtx.Unlock()

	if s.checking {
		return
	}
	logger.Debugf("Start check...")
	ticker := time.NewTicker(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	s.checkCancelFn = cancel
	s.checking = true

	go func() {
		s.tryCheck(ctx)
		for {
			select {
			case <-ticker.C:
				s.tryCheck(ctx)
			case <-ctx.Done():
				logger.Debugf("Check canceled")
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *service) stopCheck() {
	s.checkMtx.Lock()
	defer s.checkMtx.Unlock()

	logger.Debugf("Stop check...")
	s.checking = false
	s.checkCancelFn()
	// We should give it little bit of time to finish checking after the cancel
	// otherwise it might error trying to write to a closed database.
	// This wait isn't strictly required but we do it to be nice.
	// TODO: Use a WaitGroup with a timeout or channel
	for i := 0; i < 100; i++ {
		if !s.checking {
			logger.Debugf("Check stopped")
			return
		}
		time.Sleep(time.Millisecond * 10)
	}
	logger.Debugf("Timed out waiting for stop check")
}
