package service

import (
	"sync"

	"github.com/keys-pub/keys"
)

// func (s *service) watchInternalLn(e *keys.WatchEvent) {
// 	c.watchMtx.Lock()
// 	defer s.watchMtx.Unlock()
// 	c.watchLast = e
// 	c.watchLn(e)
// }

func (s *service) watchReqClose() {
	s.watchMtx.Lock()
	defer s.watchMtx.Unlock()
	s.watchLn = func(e *keys.WatchEvent) {}
	if s.watchWg != nil {
		s.watchWg.Done()
		s.watchWg = nil
	}
}

// Watch (RPC) watches for events
func (s *service) Watch(req *WatchRequest, stream Keys_WatchServer) error {
	s.watchReqClose()
	s.watchMtx.Lock()

	ln := func(event *keys.WatchEvent) {
		we := WatchEvent{
			Status: watchEventStatus(event.Status),
			Path:   event.Path,
		}

		if err := stream.Send(&we); err != nil {
			logger.Errorf("Failed to send watch event: %s", err)
		}
	}

	if s.watchLast != nil {
		ln(s.watchLast)
	}
	s.watchLn = ln
	s.watchWg = &sync.WaitGroup{}
	s.watchWg.Add(1)
	s.watchMtx.Unlock()

	s.watchWg.Wait()
	return nil
}

// SyncEventStatus converts to SyncStatus
func watchEventStatus(s keys.WatchStatus) WatchStatus {
	switch s {
	case keys.WatchStatusOutage:
		return WatchStatusOutage
	case keys.WatchStatusDisrupted:
		return WatchStatusDisrupted
	case keys.WatchStatusStarting:
		return WatchStatusStarting
	case keys.WatchStatusData:
		return WatchStatusData
	default:
		return WatchStatusUnknown
	}
}
