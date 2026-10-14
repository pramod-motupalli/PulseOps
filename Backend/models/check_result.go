package models

type CheckResult struct {
	ID              int    `json:"id"`
	MonitorID       int    `json:"monitor_id"`
	StatusCode      int    `json:"status_code"`
	ResponseTimeMS  int64  `json:"response_time_ms"`
	IsUp            bool   `json:"is_up"`
	ErrorMessage    string `json:"error_message,omitempty"`
	CheckedAt       string `json:"checked_at"`
}