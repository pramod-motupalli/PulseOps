package handlers

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var httpClient = &http.Client{
	Timeout: 6 * time.Second,
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	},
}

type MonitorHandler struct {
	DB *sql.DB
}

func NewMonitorHandler(db *sql.DB) *MonitorHandler {
	return &MonitorHandler{DB: db}
}

type NodeResponse struct {
	ID             int    `json:"id"`
	NodeID         string `json:"nodeId"`
	Name           string `json:"name"`
	Endpoint       string `json:"endpoint"`
	URL            string `json:"url"`
	Method         string `json:"method"`
	Status         string `json:"status"`
	StatusCode     int    `json:"statusCode"`
	ResponseTimeMS int64  `json:"responseTimeMs"`
	ErrorMessage   string `json:"errorMessage,omitempty"`
	LastChecked    string `json:"lastChecked,omitempty"`
	CreatedAt      string `json:"createdAt"`
}

type AddNodeRequest struct {
	Name     string `json:"name" binding:"required"`
	Endpoint string `json:"endpoint"`
	URL      string `json:"url"`
}

type TestEndpointRequest struct {
	Endpoint string `json:"endpoint"`
	URL      string `json:"url"`
}

func normalizeURL(targetURL string) string {
	clean := strings.TrimSpace(targetURL)
	if clean == "" {
		return ""
	}
	if !strings.HasPrefix(clean, "http://") && !strings.HasPrefix(clean, "https://") {
		if strings.HasPrefix(clean, "localhost") || strings.HasPrefix(clean, "127.0.0.1") {
			return "http://" + clean
		}
		return "https://" + clean
	}
	return clean
}

func classifyStatus(code, isUp int) string {
	if isUp == 1 && code >= 200 && code < 400 {
		return "Healthy"
	}
	if code >= 400 && code < 500 {
		return "Warning"
	}
	return "Critical"
}

func buildNodeResponse(id int, name, url, method, status string, code int, respTime int64, errMsg, checkedAt, createdAt string) NodeResponse {
	return NodeResponse{
		ID:             id,
		NodeID:         fmt.Sprintf("node-%02d", id),
		Name:           name,
		Endpoint:       url,
		URL:            url,
		Method:         method,
		Status:         status,
		StatusCode:     code,
		ResponseTimeMS: respTime,
		ErrorMessage:   errMsg,
		LastChecked:    checkedAt,
		CreatedAt:      createdAt,
	}
}

func parseNodeID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(strings.TrimPrefix(c.Param("id"), "node-"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return 0, false
	}
	return id, true
}

func (h *MonitorHandler) saveCheckResult(monitorID int, code int, respTime int64, isUp int, errMsg, checkedAt string) {
	_, _ = h.DB.Exec(
		"INSERT INTO check_results (monitor_id, status_code, response_time_ms, is_up, error_message, checked_at) VALUES (?, ?, ?, ?, ?, ?)",
		monitorID, code, respTime, isUp, errMsg, checkedAt,
	)
}

func checkEndpoint(targetURL string) (statusCode int, responseTimeMs int64, isUp int, status string, errMsg string) {
	cleanURL := normalizeURL(targetURL)
	if cleanURL == "" {
		return 0, 0, 0, "Critical", "Endpoint URL is empty"
	}

	req, err := http.NewRequest("GET", cleanURL, nil)
	if err != nil {
		return 0, 0, 0, "Critical", fmt.Sprintf("Invalid URL: %v", err)
	}
	req.Header.Set("User-Agent", "PulseOps-Monitor/1.0 (+https://pulseops.dev)")
	req.Header.Set("Accept", "*/*")

	start := time.Now()
	resp, err := httpClient.Do(req)

	// Fallback for HTTPS endpoints that only accept HTTP
	if err != nil && strings.HasPrefix(cleanURL, "https://") && (strings.Contains(err.Error(), "server gave HTTP response") || strings.Contains(err.Error(), "oversized record")) {
		httpURL := "http://" + strings.TrimPrefix(cleanURL, "https://")
		if retryReq, retryErr := http.NewRequest("GET", httpURL, nil); retryErr == nil {
			retryReq.Header.Set("User-Agent", "PulseOps-Monitor/1.0 (+https://pulseops.dev)")
			retryReq.Header.Set("Accept", "*/*")
			if retryResp, retryDoErr := httpClient.Do(retryReq); retryDoErr == nil {
				resp = retryResp
				err = nil
			}
		}
	}

	duration := time.Since(start).Milliseconds()
	if duration <= 0 {
		duration = 1
	}

	if err != nil {
		return 0, duration, 0, "Critical", fmt.Sprintf("Request failed: %v", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 512*1024))

	code := resp.StatusCode
	isUpVal := 0
	errStr := ""
	if code >= 200 && code < 400 {
		isUpVal = 1
	} else {
		errStr = fmt.Sprintf("HTTP %d: %s", code, http.StatusText(code))
	}

	return code, duration, isUpVal, classifyStatus(code, isUpVal), errStr
}

func (h *MonitorHandler) GetNodes(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	rows, err := h.DB.Query(`
		SELECT m.id, m.name, m.url, m.method, m.created_at,
		       COALESCE(c.status_code, 0),
		       COALESCE(c.response_time_ms, 0),
		       COALESCE(c.is_up, 0),
		       COALESCE(c.error_message, ''),
		       COALESCE(c.checked_at, m.created_at)
		FROM monitors m
		LEFT JOIN check_results c ON c.id = (
			SELECT id FROM check_results WHERE monitor_id = m.id ORDER BY checked_at DESC, id DESC LIMIT 1
		)
		WHERE m.user_id = ?
		ORDER BY m.id DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query nodes"})
		return
	}
	defer rows.Close()

	nodes := make([]NodeResponse, 0)
	for rows.Next() {
		var id int
		var name, url, method, createdAt string
		var code int
		var respTime int64
		var isUp int
		var errMsg, checkedAt string

		if err := rows.Scan(&id, &name, &url, &method, &createdAt, &code, &respTime, &isUp, &errMsg, &checkedAt); err != nil {
			continue
		}

		nodes = append(nodes, buildNodeResponse(
			id, name, url, method, classifyStatus(code, isUp), code, respTime, errMsg, checkedAt, createdAt,
		))
	}

	c.JSON(http.StatusOK, nodes)
}

func (h *MonitorHandler) AddNode(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req AddNodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Node name and endpoint URL are required"})
		return
	}

	rawURL := req.Endpoint
	if rawURL == "" {
		rawURL = req.URL
	}
	url := normalizeURL(rawURL)
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Endpoint URL is required"})
		return
	}

	statusCode, respTime, isUp, status, errMsg := checkEndpoint(url)

	res, err := h.DB.Exec(
		"INSERT INTO monitors (user_id, name, url, method, interval_seconds, timeout_seconds) VALUES (?, ?, ?, 'GET', 60, 5)",
		userID, req.Name, url,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add node: " + err.Error()})
		return
	}

	id, _ := res.LastInsertId()
	now := time.Now().Format("2006-01-02 15:04:05")
	h.saveCheckResult(int(id), statusCode, respTime, isUp, errMsg, now)

	c.JSON(http.StatusCreated, buildNodeResponse(int(id), req.Name, url, "GET", status, statusCode, respTime, errMsg, now, now))
}

func (h *MonitorHandler) TestEndpoint(c *gin.Context) {
	var req TestEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Endpoint URL is required"})
		return
	}

	rawURL := req.Endpoint
	if rawURL == "" {
		rawURL = req.URL
	}
	target := normalizeURL(rawURL)
	if target == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Endpoint URL is required"})
		return
	}

	statusCode, respTime, isUp, status, errMsg := checkEndpoint(target)
	c.JSON(http.StatusOK, gin.H{
		"statusCode":     statusCode,
		"responseTimeMs": respTime,
		"isUp":           isUp == 1,
		"status":         status,
		"errorMessage":   errMsg,
	})
}

func (h *MonitorHandler) CheckNode(c *gin.Context) {
	userID := c.GetString("userID")
	id, ok := parseNodeID(c)
	if !ok {
		return
	}

	var name, targetURL, method, createdAt string
	err := h.DB.QueryRow("SELECT name, url, method, created_at FROM monitors WHERE id = ? AND user_id = ?", id, userID).Scan(&name, &targetURL, &method, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	statusCode, respTime, isUp, status, errMsg := checkEndpoint(targetURL)
	now := time.Now().Format("2006-01-02 15:04:05")
	h.saveCheckResult(id, statusCode, respTime, isUp, errMsg, now)

	c.JSON(http.StatusOK, buildNodeResponse(id, name, targetURL, method, status, statusCode, respTime, errMsg, now, createdAt))
}

func (h *MonitorHandler) CheckAllNodes(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	rows, err := h.DB.Query("SELECT id, name, url, method, created_at FROM monitors WHERE user_id = ? ORDER BY id DESC", userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query nodes"})
		return
	}
	defer rows.Close()

	type monitorItem struct {
		id        int
		name      string
		url       string
		method    string
		createdAt string
	}
	var items []monitorItem
	for rows.Next() {
		var it monitorItem
		if err := rows.Scan(&it.id, &it.name, &it.url, &it.method, &it.createdAt); err == nil {
			items = append(items, it)
		}
	}

	results := make([]NodeResponse, 0, len(items))
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, it := range items {
		code, respTime, isUp, status, errMsg := checkEndpoint(it.url)
		h.saveCheckResult(it.id, code, respTime, isUp, errMsg, now)
		results = append(results, buildNodeResponse(it.id, it.name, it.url, it.method, status, code, respTime, errMsg, now, it.createdAt))
	}

	c.JSON(http.StatusOK, results)
}

func (h *MonitorHandler) DeleteNode(c *gin.Context) {
	userID := c.GetString("userID")
	id, ok := parseNodeID(c)
	if !ok {
		return
	}

	_, err := h.DB.Exec("DELETE FROM monitors WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete node"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Node deleted successfully", "id": c.Param("id")})
}

func (h *MonitorHandler) GetDashboardStats(c *gin.Context) {
	userID := c.GetString("userID")

	var totalNodes, healthyNodes, alertCount int
	err := h.DB.QueryRow(`
		SELECT 
			COUNT(m.id),
			COALESCE(SUM(CASE WHEN c.is_up = 1 AND c.status_code >= 200 AND c.status_code < 400 THEN 1 ELSE 0 END), 0),
			COALESCE(SUM(CASE WHEN c.id IS NOT NULL AND (c.is_up = 0 OR c.status_code >= 400) THEN 1 ELSE 0 END), 0)
		FROM monitors m
		LEFT JOIN check_results c ON c.id = (
			SELECT id FROM check_results WHERE monitor_id = m.id ORDER BY checked_at DESC, id DESC LIMIT 1
		)
		WHERE m.user_id = ?
	`, userID).Scan(&totalNodes, &healthyNodes, &alertCount)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load stats"})
		return
	}

	warningNodes := totalNodes - healthyNodes
	if warningNodes < 0 {
		warningNodes = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"total_nodes":   totalNodes,
		"active_nodes":  totalNodes,
		"healthy_nodes": healthyNodes,
		"warning_nodes": warningNodes,
		"active_alerts": alertCount,
	})
}

func (h *MonitorHandler) GetAlerts(c *gin.Context) {
	userID := c.GetString("userID")

	rows, err := h.DB.Query(`
		SELECT m.id, m.name, m.url, c.status_code, c.error_message, c.checked_at
		FROM monitors m
		JOIN check_results c ON c.id = (
			SELECT id FROM check_results WHERE monitor_id = m.id ORDER BY checked_at DESC, id DESC LIMIT 1
		)
		WHERE m.user_id = ? AND (c.is_up = 0 OR c.status_code >= 400)
		ORDER BY c.checked_at DESC, c.id DESC
	`, userID)
	if err != nil {
		c.JSON(http.StatusOK, []gin.H{})
		return
	}
	defer rows.Close()

	alerts := make([]gin.H, 0)
	for rows.Next() {
		var id int
		var name, url, checkedAt string
		var statusCode int
		var errMsg sql.NullString
		if err := rows.Scan(&id, &name, &url, &statusCode, &errMsg, &checkedAt); err != nil {
			continue
		}

		severity := "HIGH"
		if statusCode >= 500 || statusCode == 0 {
			severity = "CRITICAL"
		}

		desc := errMsg.String
		if desc == "" {
			if statusCode == 0 {
				desc = fmt.Sprintf("Connection failed / Unreachable (%s)", url)
			} else {
				desc = fmt.Sprintf("HTTP %d error on %s", statusCode, url)
			}
		}

		alerts = append(alerts, gin.H{
			"id":       fmt.Sprintf("alt-%d", id),
			"title":    fmt.Sprintf("%s - %s", name, desc),
			"severity": severity,
			"time":     checkedAt,
			"status":   "Active",
			"node_id":  id,
		})
	}

	c.JSON(http.StatusOK, alerts)
}

func (h *MonitorHandler) StartBackgroundChecker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for range ticker.C {
		rows, err := h.DB.Query("SELECT id, url FROM monitors")
		if err != nil {
			continue
		}

		type mon struct {
			id  int
			url string
		}
		var list []mon
		for rows.Next() {
			var m mon
			if err := rows.Scan(&m.id, &m.url); err == nil {
				list = append(list, m)
			}
		}
		rows.Close()

		now := time.Now().Format("2006-01-02 15:04:05")
		for _, m := range list {
			code, respTime, isUp, _, errMsg := checkEndpoint(m.url)
			h.saveCheckResult(m.id, code, respTime, isUp, errMsg, now)
		}
	}
}
