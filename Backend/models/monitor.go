package models

type Monitor struct {
	ID               int    `json:"id"`
	UserID           int    `json:"user_id"`
	Name             string `json:"name"`
	URL              string `json:"url"`
	Method           string `json:"method"`
	IntervalSeconds  int    `json:"interval_seconds"`
	TimeoutSeconds   int    `json:"timeout_seconds"`
	CreatedAt        string `json:"created_at"`
}