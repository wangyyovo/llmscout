package log

import (
	stdlog "log"
	"sync"
	"sync/atomic"
	"time"
)

type entryRepo interface {
	Insert(Entry) (int64, error)
	Query(Filter) (QueryResult, error)
	Get(int64) (*Entry, error)
	DeleteAll() error
	GetRouteNames() ([]string, error)
}

type Service struct {
	repo         entryRepo
	entries      chan Entry
	done         chan struct{}
	wg           sync.WaitGroup
	droppedCount atomic.Int64
}

func NewService(repo entryRepo) *Service {
	s := &Service{
		repo:    repo,
		entries: make(chan Entry, 100),
		done:    make(chan struct{}),
	}
	s.wg.Add(1)
	go s.worker()
	return s
}

func (s *Service) worker() {
	defer s.wg.Done()
	for {
		select {
		case entry := <-s.entries:
			entry.CreatedAt = time.Now()
			if _, err := s.repo.Insert(entry); err != nil {
				stdlog.Printf("log service: insert failed: %v", err)
			}
		case <-s.done:
			return
		}
	}
}

// Record enqueues an entry for async storage. Drops if buffer is full.
func (s *Service) Record(entry Entry) {
	select {
	case s.entries <- entry:
	default:
		s.droppedCount.Add(1)
	}
}

// Close shuts down the worker goroutine and waits for pending writes.
func (s *Service) Close() {
	close(s.done)
	s.wg.Wait()
}

// DroppedCount returns the number of entries dropped due to a full buffer.
func (s *Service) DroppedCount() int64 {
	return s.droppedCount.Load()
}

func (s *Service) Query(filter Filter) (QueryResult, error) {
	return s.repo.Query(filter)
}

func (s *Service) Get(id int64) (*Entry, error) {
	return s.repo.Get(id)
}

func (s *Service) Clear() error {
	return s.repo.DeleteAll()
}

func (s *Service) GetRouteNames() ([]string, error) {
	return s.repo.GetRouteNames()
}
