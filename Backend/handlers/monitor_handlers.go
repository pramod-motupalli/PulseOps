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
	Name       string `json:"name" binding:"required"`
	Endpoint   string `json:"endpoint"`
	URL        string `json:"url"`
	Status     string `json:"status"`
	StatusCode int    `json:"statusCode"`
}

type TestEndpointRequest struct {
	Endpoint string `json:"endpoint"`
	URL      string `json:"url"`
}

func checkEndpoint(targetURL string) (statusCode int, responseTimeMs int64, isUp int, status string, errMsg string) {
	cleanURL := strings.TrimSpace(targetURL)
	if cleanURL == "" {
		return 0, 0, 0, "Critical", "Endpoint URL is empty"
	}

	if !strings.HasPrefix(cleanURL, "http://") && !strings.HasPrefix(cleanURL, "https://") {
		if strings.HasPrefix(cleanURL, "localhost") || strings.HasPrefix(cleanURL, "127.0.0.1") {
			cleanURL = "http://" + cleanURL
		} else {
			cleanURL = "https://" + cleanURL
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   6 * time.Second,
	}

	req, err := http.NewRequest("GET", cleanURL, nil)
	if err != nil {
		return 0, 0, 0, "Critical", fmt.Sprintf("Invalid URL: %v", err)
	}
	req.Header.Set("User-Agent", "PulseOps-Monitor/1.0 (+https://pulseops.dev)")
	req.Header.Set("Accept", "*/*")

	start := time.Now()
	resp, err := client.Do(req)

	if err != nil && strings.HasPrefix(cleanURL, "https://") && (strings.Contains(err.Error(), "server gave HTTP response") || strings.Contains(err.Error(), "oversized record")) {
		httpURL := "http://" + strings.TrimPrefix(cleanURL, "https://")
		if retryReq, retryErr := http.NewRequest("GET", httpURL, nil); retryErr == nil {
			retryReq.Header.Set("User-Agent", "PulseOps-Monitor/1.0 (+https://pulseops.dev)")
			retryReq.Header.Set("Accept", "*/*")
			if retryResp, retryDoErr := client.Do(retryReq); retryDoErr == nil {
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
	if code >= 200 && code < 400 {
		return code, duration, 1, "Healthy", ""
	} else if code >= 400 && code < 500 {
		return code, duration, 0, "Warning", fmt.Sprintf("HTTP %d: %s", code, http.StatusText(code))
	} else {
		return code, duration, 0, "Critical", fmt.Sprintf("HTTP %d: %s", code, http.StatusText(code))
	}
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
		var n NodeResponse
		var isUp int
		var errMsg, checkedAt string
		if err := rows.Scan(&n.ID, &n.Name, &n.URL, &n.Method, &n.CreatedAt, &n.StatusCode, &n.ResponseTimeMS, &isUp, &errMsg, &checkedAt); err != nil {
			continue
		}
		n.Endpoint = n.URL
		n.NodeID = fmt.Sprintf("node-%02d", n.ID)
		n.ErrorMessage = errMsg
		n.LastChecked = checkedAt

		if n.StatusCode >= 200 && n.StatusCode < 400 && isUp == 1 {
			n.Status = "Healthy"
		} else if n.StatusCode >= 400 && n.StatusCode < 500 {
			n.Status = "Warning"
		} else {
			n.Status = "Critical"
		}
		nodes = append(nodes, n)
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

	url := req.Endpoint
	if url == "" {
		url = req.URL
	}
	url = strings.TrimSpace(url)
	if url == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Endpoint URL is required"})
		return
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		if strings.HasPrefix(url, "localhost") || strings.HasPrefix(url, "127.0.0.1") {
			url = "http://" + url
		} else {
			url = "https://" + url
		}
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
	_, _ = h.DB.Exec(
		"INSERT INTO check_results (monitor_id, status_code, response_time_ms, is_up, error_message, checked_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, statusCode, respTime, isUp, errMsg, now,
	)

	c.JSON(http.StatusCreated, NodeResponse{
		ID:             int(id),
		NodeID:         fmt.Sprintf("node-%02d", id),
		Name:           req.Name,
		Endpoint:       url,
		URL:            url,
		Method:         "GET",
		Status:         status,
		StatusCode:     statusCode,
		ResponseTimeMS: respTime,
		ErrorMessage:   errMsg,
		LastChecked:    now,
		CreatedAt:      now,
	})
}

func (h *MonitorHandler) TestEndpoint(c *gin.Context) {
	var req TestEndpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Endpoint URL is required"})
		return
	}

	target := req.Endpoint
	if target == "" {
		target = req.URL
	}
	target = strings.TrimSpace(target)
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
	idParam := c.Param("id")
	cleanedID := strings.TrimPrefix(idParam, "node-")
	id, err := strconv.Atoi(cleanedID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	var name, targetURL, method, createdAt string
	err = h.DB.QueryRow("SELECT name, url, method, created_at FROM monitors WHERE id = ? AND user_id = ?", id, userID).Scan(&name, &targetURL, &method, &createdAt)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Node not found"})
		return
	}

	statusCode, respTime, isUp, status, errMsg := checkEndpoint(targetURL)
	now := time.Now().Format("2006-01-02 15:04:05")
	_, _ = h.DB.Exec(
		"INSERT INTO check_results (monitor_id, status_code, response_time_ms, is_up, error_message, checked_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, statusCode, respTime, isUp, errMsg, now,
	)

	c.JSON(http.StatusOK, NodeResponse{
		ID:             id,
		NodeID:         fmt.Sprintf("node-%02d", id),
		Name:           name,
		Endpoint:       targetURL,
		URL:            targetURL,
		Method:         method,
		Status:         status,
		StatusCode:     statusCode,
		ResponseTimeMS: respTime,
		ErrorMessage:   errMsg,
		LastChecked:    now,
		CreatedAt:      createdAt,
	})
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
		statusCode, respTime, isUp, status, errMsg := checkEndpoint(it.url)
		_, _ = h.DB.Exec(
			"INSERT INTO check_results (monitor_id, status_code, response_time_ms, is_up, error_message, checked_at) VALUES (?, ?, ?, ?, ?, ?)",
			it.id, statusCode, respTime, isUp, errMsg, now,
		)
		results = append(results, NodeResponse{
			ID:             it.id,
			NodeID:         fmt.Sprintf("node-%02d", it.id),
			Name:           it.name,
			Endpoint:       it.url,
			URL:            it.url,
			Method:         it.method,
			Status:         status,
			StatusCode:     statusCode,
			ResponseTimeMS: respTime,
			ErrorMessage:   errMsg,
			LastChecked:    now,
			CreatedAt:      it.createdAt,
		})
	}

	c.JSON(http.StatusOK, results)
}

func (h *MonitorHandler) DeleteNode(c *gin.Context) {
	userID := c.GetString("userID")
	idParam := c.Param("id")

	cleanedID := strings.TrimPrefix(idParam, "node-")
	id, err := strconv.Atoi(cleanedID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid node ID"})
		return
	}

	_, err = h.DB.Exec("DELETE FROM monitors WHERE id = ? AND user_id = ?", id, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete node"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Node deleted successfully", "id": idParam})
}

func (h *MonitorHandler) GetDashboardStats(c *gin.Context) {
	userID := c.GetString("userID")

	var totalNodes int
	_ = h.DB.QueryRow("SELECT COUNT(*) FROM monitors WHERE user_id = ?", userID).Scan(&totalNodes)

	var healthyNodes int
	_ = h.DB.QueryRow(`
		SELECT COUNT(DISTINCT m.id)
		FROM monitors m
		JOIN check_results c ON c.id = (
			SELECT id FROM check_results WHERE monitor_id = m.id ORDER BY checked_at DESC, id DESC LIMIT 1
		)
		WHERE m.user_id = ? AND c.is_up = 1 AND c.status_code >= 200 AND c.status_code < 400
	`, userID).Scan(&healthyNodes)

	var alertCount int
	_ = h.DB.QueryRow(`
		SELECT COUNT(DISTINCT m.id)
		FROM monitors m
		JOIN check_results c ON c.id = (
			SELECT id FROM check_results WHERE monitor_id = m.id ORDER BY checked_at DESC, id DESC LIMIT 1
		)
		WHERE m.user_id = ? AND (c.is_up = 0 OR c.status_code >= 400)
	`, userID).Scan(&alertCount)

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
		} else if statusCode >= 400 {
			severity = "HIGH"
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
			_, _ = h.DB.Exec(
				"INSERT INTO check_results (monitor_id, status_code, response_time_ms, is_up, error_message, checked_at) VALUES (?, ?, ?, ?, ?, ?)",
				m.id, code, respTime, isUp, errMsg, now,
			)
		}
	}
}
