package log

import "time"

type Entry struct {
	ID          int64     `json:"id"`
	RouteName   string    `json:"routeName"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Protocol    string    `json:"protocol"` // "REST" | "SSE"
	StatusCode  int       `json:"statusCode"`
	LatencyMs   int64     `json:"latencyMs"`
	ReqHeaders  string    `json:"reqHeaders"`
	ReqBody     string    `json:"reqBody"`
	RespHeaders string    `json:"respHeaders"`
	RespBody    string    `json:"respBody"`
	CreatedAt   time.Time `json:"createdAt"`
}

type Filter struct {
	Keyword    string `json:"keyword"`
	RouteName  string `json:"routeName"`
	StatusCode int    `json:"statusCode"` // 0 means all
	Protocol   string `json:"protocol"`   // empty means all
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
}

type QueryResult struct {
	List  []Entry `json:"list"`
	Total int64   `json:"total"`
	Page  int     `json:"page"`
}
