package storage

import (
	"database/sql"
	"time"

	"github.com/llmscout/llmscout/internal/route"
)

type RoutesRepo struct{ db *sql.DB }

func NewRoutesRepo(db *sql.DB) *RoutesRepo {
	return &RoutesRepo{db: db}
}

func (r *RoutesRepo) List() ([]route.Rule, error) {
	rows, err := r.db.Query("SELECT id, name, type, path, target_url, created_at, updated_at FROM routes ORDER BY id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []route.Rule
	for rows.Next() {
		var rule route.Rule
		if err := rows.Scan(&rule.ID, &rule.Name, &rule.Type, &rule.Path, &rule.TargetURL, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func (r *RoutesRepo) Add(rule route.Rule) (int64, error) {
	now := time.Now()
	res, err := r.db.Exec(
		"INSERT INTO routes (name, type, path, target_url, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		rule.Name, rule.Type, rule.Path, rule.TargetURL, now, now,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *RoutesRepo) Update(id int64, rule route.Rule) error {
	_, err := r.db.Exec(
		"UPDATE routes SET name=?, type=?, path=?, target_url=?, updated_at=? WHERE id=?",
		rule.Name, rule.Type, rule.Path, rule.TargetURL, time.Now(), id,
	)
	return err
}

func (r *RoutesRepo) Delete(id int64) error {
	_, err := r.db.Exec("DELETE FROM routes WHERE id=?", id)
	return err
}

func (r *RoutesRepo) Get(id int64) (*route.Rule, error) {
	var rule route.Rule
	err := r.db.QueryRow("SELECT id, name, type, path, target_url, created_at, updated_at FROM routes WHERE id=?", id).
		Scan(&rule.ID, &rule.Name, &rule.Type, &rule.Path, &rule.TargetURL, &rule.CreatedAt, &rule.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &rule, nil
}
