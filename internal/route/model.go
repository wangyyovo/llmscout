package route

type Rule struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"` // "prefix" or "exact"
	Path      string `json:"path"`
	TargetURL string `json:"targetUrl"`
	CreatedAt int64  `json:"createdAt"` // UnixMilli
	UpdatedAt int64  `json:"updatedAt"` // UnixMilli
}
