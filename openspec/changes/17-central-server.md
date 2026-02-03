# 17: Central Server Architecture (Business Tier)

**Status:** Planning  
**Priority:** High  
**Effort:** Large (2-3 weeks)

## Problem

目前 murmur-ai 的 Team Sharing (spec 09) 基於 Git repo，適合小團隊。但對於 30+ 人的企業團隊有幾個問題：

1. **Git conflicts** — 多人同時 push patterns 會衝突
2. **無權限控制** — 每個人都有 repo full access
3. **無即時同步** — 需要手動 pull
4. **無使用統計** — 無法看到團隊的 AI 使用狀況
5. **無 audit log** — 合規需求無法滿足

需要一個中央 Server 來支援企業級的團隊協作。

## Solution

建立 `murmur-server` — 一個 Go-based 中央 server，提供：
- REST API for patterns/team management
- WebSocket for real-time sync
- PostgreSQL for persistent storage
- JWT-based authentication
- 使用統計 dashboard

## Architecture

```
murmur-server/
├── cmd/
│   └── murmur-server/
│       └── main.go           # Server entrypoint
├── internal/
│   ├── api/
│   │   ├── router.go         # Chi/Gin router setup
│   │   ├── middleware.go     # Auth, logging, CORS
│   │   ├── auth.go           # Login, register, refresh
│   │   ├── patterns.go       # CRUD patterns
│   │   ├── team.go           # Team management
│   │   ├── stats.go          # Usage statistics
│   │   └── errors.go         # Error responses
│   ├── auth/
│   │   ├── jwt.go            # JWT generation/validation
│   │   ├── password.go       # bcrypt hashing
│   │   └── claims.go         # JWT claims struct
│   ├── storage/
│   │   ├── postgres.go       # PostgreSQL connection
│   │   ├── users.go          # User CRUD
│   │   ├── teams.go          # Team CRUD
│   │   ├── patterns.go       # Pattern CRUD
│   │   └── stats.go          # Stats storage
│   └── sync/
│       ├── hub.go            # WebSocket hub
│       ├── client.go         # WebSocket client
│       └── messages.go       # Sync message types
├── migrations/
│   ├── 001_initial.up.sql
│   ├── 001_initial.down.sql
│   ├── 002_stats.up.sql
│   └── 002_stats.down.sql
├── config/
│   └── config.go             # Server config struct
├── Dockerfile
├── docker-compose.yaml
└── README.md
```

## Database Schema

### 001_initial.up.sql

```sql
-- Teams
CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'free', -- free, team, enterprise
    max_members INT NOT NULL DEFAULT 5,
    max_patterns INT NOT NULL DEFAULT 100,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Users
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    team_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'member', -- admin, member
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_team_id ON users(team_id);

-- Patterns (team-shared learning patterns)
CREATE TABLE patterns (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    content JSONB NOT NULL,  -- pattern definition
    category VARCHAR(100),   -- coding, refactor, debug, etc.
    tags TEXT[],
    author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    is_public BOOLEAN NOT NULL DEFAULT false,
    usage_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_patterns_team_id ON patterns(team_id);
CREATE INDEX idx_patterns_category ON patterns(category);
CREATE INDEX idx_patterns_tags ON patterns USING GIN(tags);

-- Refresh tokens for JWT
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
```

### 002_stats.up.sql

```sql
-- Usage statistics
CREATE TABLE stats (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    tool VARCHAR(50) NOT NULL,  -- claude, gemini, auggie, etc.
    action VARCHAR(50) NOT NULL, -- run, learn, sync
    duration_ms INT,
    tokens_used INT,
    cost_usd DECIMAL(10, 6),
    prompt_hash VARCHAR(64),  -- for dedup tracking
    pattern_id UUID REFERENCES patterns(id) ON DELETE SET NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_stats_user_id ON stats(user_id);
CREATE INDEX idx_stats_team_id ON stats(team_id);
CREATE INDEX idx_stats_tool ON stats(tool);
CREATE INDEX idx_stats_created_at ON stats(created_at);

-- Aggregated daily stats (for faster dashboard queries)
CREATE TABLE stats_daily (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    team_id UUID REFERENCES teams(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    tool VARCHAR(50) NOT NULL,
    run_count INT NOT NULL DEFAULT 0,
    total_duration_ms BIGINT NOT NULL DEFAULT 0,
    total_tokens INT NOT NULL DEFAULT 0,
    total_cost_usd DECIMAL(10, 4) NOT NULL DEFAULT 0,
    UNIQUE(user_id, team_id, date, tool)
);

CREATE INDEX idx_stats_daily_team_date ON stats_daily(team_id, date);
```

## REST API

### Authentication

```
POST /auth/register
  Request:  { "email": "...", "password": "...", "name": "..." }
  Response: { "user": {...}, "access_token": "...", "refresh_token": "..." }

POST /auth/login
  Request:  { "email": "...", "password": "..." }
  Response: { "user": {...}, "access_token": "...", "refresh_token": "..." }

POST /auth/refresh
  Request:  { "refresh_token": "..." }
  Response: { "access_token": "...", "refresh_token": "..." }

POST /auth/logout
  Request:  { "refresh_token": "..." }
  Response: { "success": true }
```

### Patterns

```
GET    /api/patterns              # List team patterns (paginated)
       Query: ?page=1&limit=20&category=coding&search=refactor
       Response: { "patterns": [...], "total": 100, "page": 1 }

GET    /api/patterns/:id          # Get single pattern
       Response: { "pattern": {...} }

POST   /api/patterns              # Create pattern
       Request:  { "name": "...", "content": {...}, "category": "...", "tags": [...] }
       Response: { "pattern": {...} }

PUT    /api/patterns/:id          # Update pattern
       Request:  { "name": "...", "content": {...} }
       Response: { "pattern": {...} }

DELETE /api/patterns/:id          # Delete pattern (admin only)
       Response: { "success": true }

POST   /api/patterns/:id/use      # Record pattern usage
       Response: { "success": true }
```

### Team Management

```
GET    /api/team                  # Get current user's team
       Response: { "team": {...}, "members": [...] }

POST   /api/team                  # Create team (becomes admin)
       Request:  { "name": "..." }
       Response: { "team": {...} }

PUT    /api/team                  # Update team (admin only)
       Request:  { "name": "..." }
       Response: { "team": {...} }

POST   /api/team/invite           # Invite member (admin only)
       Request:  { "email": "..." }
       Response: { "invite_code": "..." }

POST   /api/team/join             # Join team with invite code
       Request:  { "invite_code": "..." }
       Response: { "team": {...} }

DELETE /api/team/members/:id      # Remove member (admin only)
       Response: { "success": true }
```

### Statistics

```
GET    /api/stats/me              # My usage stats
       Query: ?from=2024-01-01&to=2024-01-31
       Response: { "stats": {...} }

GET    /api/stats/team            # Team usage stats (admin only)
       Query: ?from=2024-01-01&to=2024-01-31
       Response: { "stats": {...}, "members": [...] }

GET    /api/stats/leaderboard     # Team leaderboard
       Response: { "leaderboard": [...] }
```

### Real-time Sync

```
WebSocket /ws/sync
  Auth: ?token=<access_token>

  Client → Server:
    { "type": "subscribe", "channels": ["patterns", "stats"] }
    { "type": "pattern_created", "pattern": {...} }
    { "type": "pattern_updated", "pattern": {...} }
    { "type": "stat_recorded", "stat": {...} }

  Server → Client:
    { "type": "pattern_created", "pattern": {...}, "user": {...} }
    { "type": "pattern_updated", "pattern": {...}, "user": {...} }
    { "type": "pattern_deleted", "pattern_id": "..." }
    { "type": "member_joined", "user": {...} }
    { "type": "stats_updated", "summary": {...} }
```

## CLI Integration

### Config Changes

```go
// internal/config/config.go

type Config struct {
    // ... existing fields
    Server ServerConfig `yaml:"server"`
}

type ServerConfig struct {
    URL         string `yaml:"url"`          // e.g., "https://murmur.example.com"
    AccessToken string `yaml:"access_token"` // stored after login
    RefreshToken string `yaml:"refresh_token"`
    Mode        string `yaml:"mode"`         // "git" | "server" | "auto"
}
```

### New Commands

```bash
# Server connection
mur config server https://murmur.example.com
mur config server --clear

# Authentication
mur login                    # Interactive login
mur login --email x --password y
mur logout
mur whoami                   # Show current user

# Sync (auto-detects git vs server)
mur sync                     # If server configured, sync to server
mur sync --force-git         # Force git sync
mur sync --force-server      # Force server sync

# Team (server-only)
mur team create "My Team"    # Create team
mur team invite user@example.com
mur team join <invite-code>
mur team members
mur team stats
mur team leave

# Stats
mur stats                    # My stats
mur stats --team             # Team stats (admin)
mur stats --export csv       # Export to CSV
```

### Sync Logic

```go
// internal/sync/sync.go

func Sync() error {
    cfg, _ := config.Load()
    
    switch cfg.Server.Mode {
    case "git":
        return team.Sync()  // existing git sync
    case "server":
        return syncToServer(cfg)
    case "auto":
        if cfg.Server.URL != "" && cfg.Server.AccessToken != "" {
            return syncToServer(cfg)
        }
        if cfg.Team.Repo != "" {
            return team.Sync()
        }
        return errors.New("no sync destination configured")
    }
}

func syncToServer(cfg *config.Config) error {
    // 1. Get local patterns marked as team_shared
    // 2. POST/PUT to server
    // 3. GET server patterns not in local
    // 4. Merge into local
    // 5. Handle conflicts (server wins by default)
}
```

### Background Stat Reporting

```go
// internal/stats/reporter.go

// StartReporter starts a background goroutine that batches and reports stats
func StartReporter(cfg *config.Config) {
    // Buffer stats locally
    // Every 5 minutes (or on N events), POST to /api/stats
    // Handles offline mode (queue until online)
}
```

## Deployment

### docker-compose.yaml

```yaml
version: '3.8'

services:
  server:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgres://murmur:secret@db:5432/murmur?sslmode=disable
      - JWT_SECRET=${JWT_SECRET}
      - CORS_ORIGINS=https://murmur.example.com
    depends_on:
      - db
    restart: unless-stopped

  db:
    image: postgres:15-alpine
    environment:
      - POSTGRES_USER=murmur
      - POSTGRES_PASSWORD=secret
      - POSTGRES_DB=murmur
    volumes:
      - pgdata:/var/lib/postgresql/data
    restart: unless-stopped

  migrate:
    build: .
    command: /app/murmur-server migrate up
    environment:
      - DATABASE_URL=postgres://murmur:secret@db:5432/murmur?sslmode=disable
    depends_on:
      - db

volumes:
  pgdata:
```

### Environment Variables

```bash
# Required
DATABASE_URL=postgres://user:pass@host:5432/murmur?sslmode=require
JWT_SECRET=your-256-bit-secret

# Optional
PORT=8080
CORS_ORIGINS=https://murmur.example.com,https://app.murmur.io
LOG_LEVEL=info
```

## Security Considerations

1. **JWT tokens** — Access token expires in 15 minutes, refresh in 7 days
2. **Password hashing** — bcrypt with cost 12
3. **HTTPS only** — No HTTP in production
4. **Rate limiting** — 100 req/min per user, 1000 req/min per team
5. **Input validation** — All inputs sanitized
6. **Audit log** — All sensitive operations logged

## Pricing Tiers (Reference)

| Plan | Members | Patterns | Stats Retention | Price |
|------|---------|----------|-----------------|-------|
| Free | 5 | 100 | 7 days | $0 |
| Team | 30 | 1,000 | 90 days | $10/user/mo |
| Enterprise | Unlimited | Unlimited | 1 year | Contact |

## Migration from Git Sync

For teams migrating from git-based sync (spec 09):

```bash
# 1. Setup server
mur config server https://murmur.example.com
mur login

# 2. Create team on server
mur team create "My Team"

# 3. Import existing patterns
mur sync --import-from-git

# 4. Invite members
mur team invite alice@example.com
mur team invite bob@example.com

# 5. Switch to server mode
mur config set server.mode server
```

## Acceptance Criteria

- [ ] Database schema defined and migrations created
- [ ] REST API endpoints documented with request/response formats
- [ ] WebSocket sync protocol defined
- [ ] CLI commands documented with examples
- [ ] Docker deployment config ready
- [ ] Security considerations addressed
- [ ] Migration path from git sync documented

## Implementation Phases

### Phase 1: Core Server (Week 1)
- Database setup + migrations
- User auth (register/login/JWT)
- Basic patterns CRUD

### Phase 2: Team Features (Week 2)
- Team creation/management
- Invite system
- Role-based permissions

### Phase 3: Real-time + Stats (Week 3)
- WebSocket sync
- Usage statistics
- Dashboard queries

### Phase 4: CLI Integration (Week 4)
- `mur login/logout`
- `mur sync` server mode
- `mur team` commands
- Background stat reporting

## Dependencies

- `github.com/go-chi/chi/v5` — HTTP router
- `github.com/golang-jwt/jwt/v5` — JWT handling
- `github.com/jackc/pgx/v5` — PostgreSQL driver
- `github.com/gorilla/websocket` — WebSocket
- `github.com/golang-migrate/migrate/v4` — DB migrations
- `golang.org/x/crypto` — bcrypt

## Related Specs

- `09-team-sharing.md` — Git-based team sync (this replaces for large teams)
- `10-stats.md` — Local stats (this extends to server-side)
- `11-learning-repo.md` — Learning patterns storage

## Notes

這是規劃文件，不是完整實作。實際開發時需要：
1. 設定測試環境 (testcontainers for PostgreSQL)
2. 撰寫 API 測試
3. 建立 CI/CD pipeline
4. 設定 production 部署 (可考慮 fly.io, Railway, 或 self-hosted)
