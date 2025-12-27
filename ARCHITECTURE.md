# Architecture

This document describes the architecture of the MC Admin web application, a Go-based tool for managing Minecraft servers via RCON.

## Overview

```mermaid
graph TB
    subgraph Frontend
        Browser[Browser]
        HTMX[HTMX]
        Templates[HTML Templates]
    end

    subgraph API Layer
        Gin[Gin Web Framework]
        Handlers[HTTP Handlers]
        Auth[Discord OAuth]
    end

    subgraph Service Layer
        ServerSvc[ServerService]
        WhitelistSvc[WhitelistService]
        CommandSvc[CommandService]
        WorldSvc[WorldService]
        FileSvc[FileService]
    end

    subgraph RCON Layer
        RconClient[MinecraftRconClient]
    end

    subgraph External
        MC[Minecraft Server]
        Discord[Discord OAuth]
        Ashcon[Ashcon API]
        FileSystem[File System]
    end

    Browser --> HTMX
    HTMX --> Gin
    Gin --> Auth
    Auth --> Discord
    Gin --> Handlers
    Handlers --> Templates
    Handlers --> ServerSvc
    Handlers --> WhitelistSvc
    Handlers --> CommandSvc
    Handlers --> WorldSvc
    Handlers --> FileSvc
    ServerSvc --> RconClient
    WhitelistSvc --> RconClient
    WhitelistSvc --> Ashcon
    CommandSvc --> RconClient
    WorldSvc --> RconClient
    FileSvc --> FileSystem
    RconClient --> MC
```

## Three-Layer Architecture

The codebase follows a clean three-layer architecture pattern with clear separation of concerns:

```mermaid
graph LR
    subgraph "1. API Layer"
        direction TB
        A1[HTTP Handlers]
        A2[Route Definitions]
        A3[Template Rendering]
    end

    subgraph "2. Service Layer"
        direction TB
        S1[Business Logic]
        S2[Data Transformation]
        S3[Response Parsing]
    end

    subgraph "3. RCON Layer"
        direction TB
        R1[Protocol Handling]
        R2[Connection Management]
        R3[Command Execution]
    end

    A1 --> S1
    S1 --> R1
```

### Layer Responsibilities

| Layer | Location | Responsibility |
|-------|----------|----------------|
| **API** | `internal/api/` | HTTP routing, request handling, authentication, template rendering |
| **Service** | `internal/services/` | Business logic, data transformation, RCON response parsing |
| **RCON** | `internal/rcon/` | Low-level RCON protocol communication, connection management |

## Directory Structure

```
mc-admin/
├── main.go                     # Application entry point
├── internal/
│   ├── rcon/                   # RCON layer
│   │   └── client.go           # MinecraftRconClient
│   ├── services/               # Service layer
│   │   ├── server.go           # Player info
│   │   ├── command.go          # Raw commands
│   │   ├── whitelist.go        # Whitelist management
│   │   ├── world.go            # World/time operations
│   │   ├── files.go            # File operations
│   │   └── gametime.go         # Game time parsing
│   ├── api/                    # API layer
│   │   ├── server.go           # Route initialization
│   │   ├── player.go           # Player handlers
│   │   ├── command.go          # Command console
│   │   ├── whitelist.go        # Whitelist handlers
│   │   ├── world.go            # World handlers
│   │   ├── files.go            # File handlers
│   │   └── auth.go             # Discord OAuth
│   ├── clients/                # External API clients
│   │   └── ashcon.go           # Mojang username verification
│   ├── files/                  # File system abstraction
│   │   └── client.go           # MinecraftFilesClient
│   ├── config/                 # Configuration
│   │   └── environment.go      # Environment variables
│   └── utils/                  # Utilities
│       ├── strings.go
│       └── htmx.go
├── templates/                  # HTML templates
└── static/                     # CSS, images
```

## Dependency Flow

```mermaid
graph TD
    main[main.go]

    subgraph Initialization
        env[Load .env]
        rcon[Create RconClient]
        web[Initialize WebServer]
    end

    subgraph Services
        ss[ServerService]
        cs[CommandService]
        ws[WhitelistService]
        wos[WorldService]
        fs[FileService]
    end

    subgraph Handlers
        ph[Player Handlers]
        ch[Command Handlers]
        wh[Whitelist Handlers]
        woh[World Handlers]
        fh[File Handlers]
    end

    main --> env
    env --> rcon
    rcon --> web
    web --> ss
    web --> cs
    web --> ws
    web --> wos
    web --> fs
    ss --> ph
    cs --> ch
    ws --> wh
    wos --> woh
    fs --> fh
```

Dependencies are injected from `main.go` downward:

1. **main.go** creates the `MinecraftRconClient` from environment variables
2. **api.InitializeWebServer()** receives the client and creates all services
3. **Services** receive the RCON client via constructor injection
4. **Handlers** are factory functions that close over service dependencies

## Request Flow

```mermaid
sequenceDiagram
    participant B as Browser
    participant H as HTMX
    participant G as Gin Router
    participant A as Auth Middleware
    participant Ha as Handler
    participant S as Service
    participant R as RconClient
    participant M as Minecraft

    B->>H: User clicks button
    H->>G: HTTP GET /server-info
    G->>A: Check authentication
    A->>Ha: Pass to handler
    Ha->>S: GetServerPlayerInfo()
    S->>R: ExecuteCommand("list")
    R->>M: RCON: list
    M-->>R: "There are 2/20 players: Alice, Bob"
    R-->>S: Raw response
    S-->>Ha: ServerPlayerInfo struct
    Ha-->>G: Render player_list.html
    G-->>H: HTML partial
    H->>B: Swap into DOM
```

## Frontend Architecture

The frontend uses server-side rendering with HTMX for dynamic updates without JavaScript frameworks.

```mermaid
graph TB
    subgraph Browser
        index[index.html - Shell]
        modal[Modal Container]
        subpage[Subpage Panel]
    end

    subgraph "HTMX Partials"
        pl[player_list.html]
        wl[whitelist.html]
        wo[world.html]
        cc[command_console.html]
        fi[files.html]
    end

    subgraph "Sub-partials"
        ws[world_stats.html]
        cv[clock_view.html]
        ce[clock_edit.html]
        cr[command_result.html]
    end

    index --> modal
    index --> subpage
    subpage -.-> pl
    subpage -.-> wl
    subpage -.-> wo
    subpage -.-> cc
    subpage -.-> fi
    wo --> ws
    wo --> cv
    wo --> ce
    cc --> cr

    style pl fill:#e1f5fe
    style wl fill:#e1f5fe
    style wo fill:#e1f5fe
    style cc fill:#e1f5fe
    style fi fill:#e1f5fe
```

### HTMX Pattern

```html
<!-- Button triggers partial replacement -->
<button hx-get="/whitelist"
        hx-target="#subpage-panel"
        hx-swap="innerHTML">
  Whitelist
</button>

<!-- Panel content is swapped dynamically -->
<div id="subpage-panel">
  <!-- Partial HTML loaded here -->
</div>
```

## Key Interfaces

```mermaid
classDiagram
    class CommandExecutor {
        <<interface>>
        +ExecuteCommand(cmd string) string, error
    }

    class FileSystemAccessor {
        <<interface>>
        +ListFiles(path string) []FileInfo, error
        +ReadFile(path string) string, error
        +SaveFile(path string, content string) error
        +Delete(path string) error
    }

    class MojangUserNameChecker {
        <<interface>>
        +CheckMojangUsernameExists(name string) bool, error
    }

    class MinecraftRconClient {
        -host string
        -port string
        -password string
        -conn *rcon.Conn
        -mu sync.Mutex
        +ExecuteCommand(cmd string) string, error
        +Close() error
    }

    class MinecraftFilesClient {
        -basePath string
        +ListFiles(path string) []FileInfo, error
        +ReadFile(path string) string, error
        +SaveFile(path, content string) error
    }

    class AshconClient {
        -httpClient *http.Client
        +CheckMojangUsernameExists(name string) bool, error
    }

    CommandExecutor <|.. MinecraftRconClient
    FileSystemAccessor <|.. MinecraftFilesClient
    MojangUserNameChecker <|.. AshconClient
```

## Service Dependencies

```mermaid
graph LR
    subgraph Interfaces
        CE[CommandExecutor]
        FSA[FileSystemAccessor]
        MUC[MojangUserNameChecker]
    end

    subgraph Services
        SS[ServerService]
        CS[CommandService]
        WS[WhitelistService]
        WOS[WorldService]
        FS[FileService]
    end

    SS --> CE
    CS --> CE
    WS --> CE
    WS --> FSA
    WS --> MUC
    WOS --> CE
    FS --> FSA
```

## HTTP Routes

| Method | Route | Handler | Description |
|--------|-------|---------|-------------|
| GET | `/` | Index | Main dashboard |
| GET | `/server-info` | GetServerInfo | Player list (HTMX partial) |
| GET | `/whitelist` | GetWhitelist | Whitelist management |
| POST | `/whitelist/toggle` | ToggleWhitelist | Enable/disable whitelist |
| POST | `/whitelist/player` | AddWhitelistPlayer | Add player to whitelist |
| DELETE | `/whitelist/player/:name` | RemoveWhitelistPlayer | Remove player |
| GET | `/players/:name/kick` | GetKickPlayer | Kick confirmation dialog |
| POST | `/players/:name/kick` | KickPlayer | Execute kick |
| GET | `/rcon` | GetCommandConsole | RCON console |
| POST | `/commands/execute` | ExecuteRawCommand | Run RCON command |
| GET | `/world/stats` | GetWorldStats | World statistics |
| GET | `/world/clock` | GetClock | Time display |
| POST | `/world/time` | SetTime | Set game time |
| GET | `/files` | GetFiles | File browser |
| POST | `/files/upload` | UploadFile | Upload file |
| DELETE | `/files/delete` | DeleteFile | Delete file |

## Authentication Flow

```mermaid
sequenceDiagram
    participant U as User
    participant A as App
    participant D as Discord

    U->>A: GET /login
    A->>U: Redirect to Discord OAuth
    U->>D: Authorize app
    D->>A: Callback with code
    A->>D: Exchange code for token
    D->>A: Access token + user info
    A->>A: Validate user is whitelisted
    A->>U: Set session cookie
    U->>A: Subsequent requests with session
```

## Adding New Features

### Adding a New RCON Command

```mermaid
graph TD
    A[1. Add Service Method] --> B[2. Create Handler]
    B --> C[3. Register Route]
    C --> D[4. Create Template]

    subgraph "internal/services/"
        A1[Parse RCON response]
        A2[Return structured data]
    end

    subgraph "internal/api/"
        B1[Handle HTTP request]
        B2[Call service]
        B3[Render template]
    end

    subgraph "templates/"
        D1[HTML partial for HTMX]
    end

    A --> A1
    A1 --> A2
    B --> B1
    B1 --> B2
    B2 --> B3
    D --> D1
```

### Example: Adding a "Ban Player" feature

1. **Service** (`internal/services/server.go`):
```go
func (s *ServerService) BanPlayer(name, reason string) error {
    cmd := fmt.Sprintf("ban %s %s", name, reason)
    _, err := s.rconClient.ExecuteCommand(cmd)
    return err
}
```

2. **Handler** (`internal/api/player.go`):
```go
func handleBanPlayer(serverService *services.ServerService) gin.HandlerFunc {
    return func(c *gin.Context) {
        name := c.Param("name")
        reason := c.PostForm("reason")
        if err := serverService.BanPlayer(name, reason); err != nil {
            // Handle error
        }
        c.HTML(200, "ban_success.html", nil)
    }
}
```

3. **Route** (`internal/api/server.go`):
```go
protected.POST("/players/:name/ban", handleBanPlayer(parts.ServerService))
```

4. **Template** (`templates/ban_success.html`):
```html
<div class="toast success">Player banned successfully</div>
```

## Configuration

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `RCON_PASSWORD` | Yes | - | RCON password |
| `RCON_HOST` | No | localhost | Minecraft server host |
| `RCON_PORT` | No | 25575 | RCON port |
| `DISCORD_CLIENT_ID` | Yes | - | Discord OAuth client ID |
| `DISCORD_CLIENT_SECRET` | Yes | - | Discord OAuth secret |
| `SESSION_SECRET` | Yes | - | Session encryption key |

### Kubernetes Deployment

For Kubernetes environments, use `start-with-k8s.sh`:

| Variable | Default | Description |
|----------|---------|-------------|
| `K8S_NAMESPACE` | mc-red | Kubernetes namespace |
| `K8S_POD_SELECTOR` | app=mc-red-minecraft | Pod label selector |
| `LOCAL_PORT` | 25575 | Local port for forwarding |
