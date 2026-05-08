package log

import "time"

type entryRepo interface {
	Insert(Entry) (int64, error)
	Query(Filter) (QueryResult, error)
	Get(int64) (*Entry, error)
	DeleteAll() error
	GetRouteNames() ([]string, error)
}

type Service struct {
	repo    entryRepo
	entries chan Entry
}

func NewService(repo entryRepo) *Service {
	s := &Service{
		repo:    repo,
		entries: make(chan Entry, 100),
	}
	go s.worker()
	return s
}

func (s *Service) worker() {
	for entry := range s.entries {
		entry.CreatedAt = time.Now()
		_, err := s.repo.Insert(entry)
		if err != nil {
			// Log and drop — never block the proxy
		}
	}
}

// Record enqueues an entry for async storage. Drops if buffer is full.
func (s *Service) Record(entry Entry) {
	select {
	case s.entries <- entry:
	default:
	}
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
