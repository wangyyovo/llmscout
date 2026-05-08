package route

import (
	"fmt"
	"net/url"
	"strings"
	"sync"
)

type Manager struct {
	mu    sync.RWMutex
	rules []Rule
	repo  interface {
		List() ([]Rule, error)
		Add(Rule) (int64, error)
		Update(int64, Rule) error
		Delete(int64) error
		Get(int64) (*Rule, error)
	}
}

func NewManager(repo interface {
	List() ([]Rule, error)
	Add(Rule) (int64, error)
	Update(int64, Rule) error
	Delete(int64) error
	Get(int64) (*Rule, error)
}) *Manager {
	return &Manager{repo: repo}
}

func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	rules, err := m.repo.List()
	if err != nil {
		return err
	}
	m.rules = rules
	return nil
}

func (m *Manager) List() []Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Rule, len(m.rules))
	copy(result, m.rules)
	return result
}

func (m *Manager) Add(rule Rule) (int64, error) {
	id, err := m.repo.Add(rule)
	if err != nil {
		return 0, err
	}
	rule.ID = id
	m.mu.Lock()
	m.rules = append(m.rules, rule)
	m.mu.Unlock()
	return id, nil
}

func (m *Manager) Update(id int64, rule Rule) error {
	if err := m.repo.Update(id, rule); err != nil {
		return err
	}
	rule.ID = id
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, r := range m.rules {
		if r.ID == id {
			m.rules[i] = rule
			return nil
		}
	}
	return fmt.Errorf("route %d not found", id)
}

func (m *Manager) Delete(id int64) error {
	if err := m.repo.Delete(id); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, r := range m.rules {
		if r.ID == id {
			m.rules = append(m.rules[:i], m.rules[i+1:]...)
			return nil
		}
	}
	return nil
}

// Match returns the target URL and whether a match was found.
// For prefix rules, strips the path prefix and appends remaining path to target URL.
// For exact rules, returns the full target URL as-is.
func (m *Manager) Match(requestPath string) (targetURL string, routeName string, ok bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, rule := range m.rules {
		switch rule.Type {
		case "exact":
			if requestPath == rule.Path {
				return rule.TargetURL, rule.Name, true
			}
		case "prefix":
			if strings.HasPrefix(requestPath, rule.Path) {
				suffix := strings.TrimPrefix(requestPath, rule.Path)
				return strings.TrimRight(rule.TargetURL, "/") + "/" + strings.TrimLeft(suffix, "/"), rule.Name, true
			}
		}
	}
	return "", "", false
}

// IsValidURL checks that a target URL has a valid scheme and host.
func IsValidURL(target string) bool {
	u, err := url.Parse(target)
	return err == nil && u.Scheme != "" && u.Host != ""
}
