# PulseOps Backend

PulseOps Backend is a high-performance RESTful monitoring engine built in Go. It actively pings API endpoints and web services, tracks response latencies, verifies HTTP status codes, manages operational incidents, and serves real-time health metrics to the PulseOps dashboard.

---

## ⚡ Tech Stack & Architecture

- **Language:** Go (1.26+)
- **HTTP Engine:** [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database:** SQLite (`database/db` via pure-Go [`modernc.org/sqlite`](https://gitlab.com/cznic/sqlite), no CGO required)
- **Security:** JWT (`golang-jwt/jwt/v5`) & Bcrypt password hashing
- **Concurrency:** Native Go goroutines & tickers for background health checks and parallel network I/O

---

## 🚀 Core Engine Capabilities

1. **Active HTTP Verification**
   - When an endpoint is submitted or re-checked, the backend sends a live HTTP request with a 6-second timeout.
   - Measures exact round-trip latency in milliseconds.
   - Follows redirects and supports smart protocol negotiation (falls back from HTTPS to HTTP for local services like `localhost` or `127.0.0.1`).
   - Ignores self-signed certificate rejections to support internal services and staging environments.

2. **Automated Status Resolution**
   - **`Healthy` (Operational):** HTTP `200`–`399` (e.g. 200 OK, 201 Created, 204 No Content, 301/302 Redirect).
   - **`Warning` (Degraded):** HTTP `400`–`499` (e.g. 401 Unauthorized, 403 Forbidden, 404 Not Found).
   - **`Critical` (Outage / Down):** HTTP `500`–`599` (500 Internal Server Error, 502 Bad Gateway, 503 Unavailable) or `0` (connection refused, DNS resolution failure, network timeout).

3. **Background Monitoring Worker**
   - A background ticker goroutine runs periodically (every 45 seconds), executing live health checks across all configured monitors and recording timestamped audit rows in `check_results`.

4. **Incident & Alert Generation**
   - Endpoints experiencing 4xx or 5xx codes or connection failures automatically generate alerts viewable in the Incident Monitor, categorized by severity (`HIGH` or `CRITICAL`).

---

## 📁 Directory Structure

```
Backend/
├── database/
│   ├── database.go         # SQLite initialization, migrations & foreign key enforcement
│   └── db                  # Embedded SQLite database file
├── handlers/
│   ├── auth_handlers.go    # User registration, login & profile lookup
│   └── monitor_handlers.go # Node creation, live HTTP testing, re-checks & stats
├── middleware/
│   └── auth.go             # Bearer JWT verification middleware
├── models/
│   ├── check_result.go     # CheckResult data models
│   ├── monitor.go          # Monitor data models
│   └── user.go             # User registration & auth request models
├── .env                    # Environment configuration
├── go.mod                  # Go module definition
├── go.sum                  # Dependency checksums
└── main.go                 # Router initialization, CORS, routes & server entrypoint
```

---

## ⚙️ Environment Variables

Create or update `.env` in the `Backend/` directory:

```env
PORT=5000
JWT_SECRET=super_secret_pulseops_jwt_signing_key_32chars_min
```

| Variable | Description | Default |
| :--- | :--- | :--- |
| `PORT` | Listening port for the HTTP server | `5000` |
| `JWT_SECRET` | Signing secret for HS256 tokens (min 32 characters) | *Required* |

---

## 📡 API Endpoints

### 1. Health & Public Routes

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Server health check (`{"service":"PulseOps","status":"healthy"}`) |

### 2. Authentication

| Method | Endpoint | Description | Request Body |
| :--- | :--- | :--- | :--- |
| `POST` | `/register` | Register a new operator account | `{"name":"...","email":"...","password":"..."}` |
| `POST` | `/login` | Authenticate and obtain JWT token | `{"email":"...","password":"..."}` |

### 3. Services & Endpoint Monitoring *(Requires Bearer Token)*

| Method | Endpoint | Description | Request Body |
| :--- | :--- | :--- | :--- |
| `GET` | `/me` | Get profile for currently authenticated user | None |
| `GET` | `/nodes` | List all monitored nodes with latest health status & latency | None |
| `POST` | `/nodes` | Add a new endpoint (immediately pings and records real status) | `{"name":"Production API","endpoint":"https://api.example.com"}` |
| `POST` | `/nodes/test` | Test any API endpoint on-the-fly without saving | `{"endpoint":"https://api.example.com/health"}` |
| `POST` | `/nodes/:id/check` | Trigger an immediate live health check on a specific node | None |
| `POST` | `/nodes/check-all` | Trigger immediate live health checks across all active nodes | None |
| `DELETE` | `/nodes/:id` | Remove a monitored endpoint | None |

### 4. Metrics & Alerts *(Requires Bearer Token)*

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/dashboard/stats` | Cluster summary counters (`total_nodes`, `healthy_nodes`, `warning_nodes`, `active_alerts`) |
| `GET` | `/alerts` | Active incident log for degraded or down endpoints |

---

## 💻 Running the Backend

### Start in Development Mode

```bash
cd Backend
go run .
```

### Build Binary

```bash
cd Backend
go build -o pulseops-backend.exe .
./pulseops-backend.exe
```
