package storage

import (
	"database/sql"
	"fmt"
	stdlog "log"
	"strings"
	"time"

	"github.com/llmscout/llmscout/internal/log"
)

type LogsRepo struct{ db *sql.DB }

func NewLogsRepo(db *sql.DB) *LogsRepo {
	return &LogsRepo{db: db}
}

func (r *LogsRepo) Insert(entry log.Entry) (int64, error) {
	createdAt := time.UnixMilli(entry.CreatedAt)
	res, err := r.db.Exec(
		`INSERT INTO logs (route_name, method, path, protocol, status_code, latency_ms,
		 req_headers, req_body, resp_headers, resp_body, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		entry.RouteName, entry.Method, entry.Path, entry.Protocol,
		entry.StatusCode, entry.LatencyMs, entry.ReqHeaders, entry.ReqBody,
		entry.RespHeaders, entry.RespBody, createdAt,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (r *LogsRepo) Query(filter log.Filter) (log.QueryResult, error) {
	var where []string
	var args []interface{}

	if filter.RouteName != "" {
		where = append(where, "route_name = ?")
		args = append(args, filter.RouteName)
	}
	if filter.StatusCode > 0 {
		where = append(where, "status_code >= ? AND status_code < ?")
		args = append(args, filter.StatusCode, filter.StatusCode+100)
	}
	if filter.Protocol != "" {
		where = append(where, "protocol = ?")
		args = append(args, filter.Protocol)
	}
	if filter.StartTime != "" {
		where = append(where, "created_at >= ?")
		args = append(args, filter.StartTime)
	}
	if filter.EndTime != "" {
		where = append(where, "created_at <= ?")
		args = append(args, filter.EndTime)
	}
	if filter.Keyword != "" {
		where = append(where, "(req_body LIKE ? OR resp_body LIKE ?)")
		kw := "%" + filter.Keyword + "%"
		args = append(args, kw, kw)
	}

	clause := ""
	if len(where) > 0 {
		clause = " WHERE " + strings.Join(where, " AND ")
	}

	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM logs" + clause
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return log.QueryResult{}, err
	}

	offset := (filter.Page - 1) * filter.PageSize
	dataQuery := fmt.Sprintf("SELECT id, route_name, method, path, protocol, status_code, latency_ms, req_headers, req_body, resp_headers, resp_body, created_at FROM logs%s ORDER BY id DESC LIMIT ? OFFSET ?", clause)
	dataArgs := append(args, filter.PageSize, offset)
	rows, err := r.db.Query(dataQuery, dataArgs...)
	if err != nil {
		return log.QueryResult{}, err
	}
	defer rows.Close()

	var entries []log.Entry
	for rows.Next() {
		var e log.Entry
		var createdAt time.Time
		if err := rows.Scan(&e.ID, &e.RouteName, &e.Method, &e.Path, &e.Protocol,
			&e.StatusCode, &e.LatencyMs, &e.ReqHeaders, &e.ReqBody,
			&e.RespHeaders, &e.RespBody, &createdAt); err != nil {
			stdlog.Printf("scan log row: %v", err)
			continue
		}
		e.CreatedAt = createdAt.UnixMilli()
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return log.QueryResult{}, err
	}
	return log.QueryResult{List: entries, Total: total, Page: filter.Page}, nil
}

func (r *LogsRepo) Get(id int64) (*log.Entry, error) {
	var e log.Entry
	var createdAt time.Time
	err := r.db.QueryRow(
		"SELECT id, route_name, method, path, protocol, status_code, latency_ms, req_headers, req_body, resp_headers, resp_body, created_at FROM logs WHERE id=?", id,
	).Scan(&e.ID, &e.RouteName, &e.Method, &e.Path, &e.Protocol,
		&e.StatusCode, &e.LatencyMs, &e.ReqHeaders, &e.ReqBody,
		&e.RespHeaders, &e.RespBody, &createdAt)
	if err != nil {
		return nil, err
	}
	e.CreatedAt = createdAt.UnixMilli()
	return &e, nil
}

func (r *LogsRepo) DeleteAll() error {
	_, err := r.db.Exec("DELETE FROM logs")
	return err
}

func (r *LogsRepo) GetRouteNames() ([]string, error) {
	rows, err := r.db.Query("SELECT DISTINCT route_name FROM logs ORDER BY route_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return names, nil
}
