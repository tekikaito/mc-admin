# mc-admin

A web-based administration panel for Minecraft servers that provides real-time player monitoring, whitelist management, and RCON command execution through a clean, modern interface.

## Features

- **Real-time Player Monitoring**: View currently online players with auto-refresh
- **Whitelist Management**: Add and remove players from the server whitelist with Mojang username validation
- **Player Actions**: Kick players directly from the web interface
- **RCON Console**: Execute raw RCON commands with syntax highlighting
- **Discord OAuth Authentication**: Secure access control via Discord login
- **Server Information Display**: Customizable server name, version, and description
- **HTMX-Powered UI**: Dynamic updates without page refreshes
- **Graceful Shutdown**: Proper cleanup of connections and resources

## Architecture

The application follows a clean three-layer architecture:

1. **Protocol Layer** (`internal/rcon/`): Low-level RCON protocol communication using the `github.com/gorcon/rcon` library
2. **Service Layer** (`internal/services/`): Business logic and data transformation
3. **API Layer** (`internal/api/`): HTTP handlers and routing using the Gin framework

All HTML is rendered server-side with templates located in `templates/`, and HTMX is used for dynamic content updates.

## Prerequisites

- Go 1.25.4 or later
- A Minecraft server with RCON enabled
- Discord application credentials (for authentication)
- Optional: Kubernetes cluster (for remote server access)

## Quick Start

### 1. Clone the repository

```bash
git clone <repository-url>
cd mc-admin
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Configure environment variables

Create a `.env` file in the project root:

```bash
# Required: RCON Configuration
RCON_PASSWORD=your-rcon-password

# Optional: RCON Connection (defaults shown)
RCON_HOST=localhost
RCON_PORT=25575

# Required: Discord OAuth
DISCORD_CLIENT_ID=your-client-id
DISCORD_CLIENT_SECRET=your-client-secret
DISCORD_REDIRECT_URI=http://localhost:8080/auth/discord/callback
SESSION_SECRET=random-long-string-for-session-encryption

# Optional: Discord user allowlist (comma-separated Discord user IDs)
DISCORD_ALLOWED_USER_IDS=123456789,987654321

# Optional: Server Display Information
SERVER_NAME=My Minecraft Server
SERVER_HOST=play.example.com
GAME_PORT=25565
SERVER_VERSION=1.20.1
SERVER_DESCRIPTION=Welcome to our community server!
```

### 4. Run the application

**Option A: Local Development**

```bash
go run main.go
```

**Option B: With Kubernetes Port-Forwarding**

If your Minecraft server is running in Kubernetes:

```bash
# Set optional environment variables for Kubernetes
export K8S_NAMESPACE=minecraft
export K8S_POD_SELECTOR=app=mc-server
export LOCAL_PORT=25575

# Run the script (automatically sets up port-forwarding and hot-reload)
./start-with-k8s.sh
```

### 5. Access the application

Open your browser to `http://localhost:8080` and authenticate with Discord.

## Discord OAuth Setup

1. Create an application in the [Discord Developer Portal](https://discord.com/developers/applications)
2. In OAuth2 settings, add a redirect URL:
   - Local development: `http://localhost:8080/auth/discord/callback`
   - Production: `https://your-domain/auth/discord/callback`
3. Copy the Client ID and Client Secret to your `.env` file
4. Optionally, restrict access by adding Discord user IDs to `DISCORD_ALLOWED_USER_IDS`

All routes except `/auth/*` require a valid Discord session. If `DISCORD_ALLOWED_USER_IDS` is set, only those users can sign in.

## Environment Variables Reference

### Required Variables

| Variable                | Description                                                  |
| ----------------------- | ------------------------------------------------------------ |
| `RCON_PASSWORD`         | RCON password for your Minecraft server                      |
| `DISCORD_CLIENT_ID`     | Discord OAuth application client ID                          |
| `DISCORD_CLIENT_SECRET` | Discord OAuth application client secret                      |
| `DISCORD_REDIRECT_URI`  | OAuth callback URL                                           |
| `SESSION_SECRET`        | Secret key for session encryption (use a long random string) |

### Optional Variables

| Variable                   | Default                          | Description                                                          |
| -------------------------- | -------------------------------- | -------------------------------------------------------------------- |
| `RCON_HOST`                | `localhost`                      | Minecraft server hostname or IP                                      |
| `RCON_PORT`                | `25575`                          | RCON port on the Minecraft server                                    |
| `DISCORD_ALLOWED_USER_IDS` | (none)                           | Comma-separated list of Discord user IDs allowed to access the panel |
| `SERVER_NAME`              | `Minecraft Server`               | Display name shown in the UI                                         |
| `SERVER_HOST`              | `localhost`                      | Public server address shown to users                                 |
| `GAME_PORT`                | `25565`                          | Minecraft game port shown to users                                   |
| `SERVER_VERSION`           | `Unknown Version`                | Server version displayed in the UI                                   |
| `SERVER_DESCRIPTION`       | `Live status for your community` | Server description text                                              |

### Kubernetes Script Variables

| Variable           | Default                | Description                                       |
| ------------------ | ---------------------- | ------------------------------------------------- |
| `K8S_NAMESPACE`    | `mc-red`               | Kubernetes namespace containing the Minecraft pod |
| `K8S_POD_SELECTOR` | `app=mc-red-minecraft` | Label selector to find the Minecraft pod          |
| `LOCAL_PORT`       | `25575`                | Local port for port-forwarding                    |

## Docker Deployment

Build and run using Docker:

```bash
# Build the image
docker build -t mc-admin .

# Run the container
docker run -p 8080:8080 \
  -e RCON_PASSWORD=your-password \
  -e RCON_HOST=your-minecraft-server \
  -e DISCORD_CLIENT_ID=your-client-id \
  -e DISCORD_CLIENT_SECRET=your-client-secret \
  -e DISCORD_REDIRECT_URI=http://localhost:8080/auth/discord/callback \
  -e SESSION_SECRET=your-session-secret \
  mc-admin
```

## Development

### Running Tests

```bash
go test ./...
```

### Hot Reload Development

The `start-with-k8s.sh` script includes hot-reload functionality using `reflex`:

```bash
# Reflex will automatically restart the server when .go or .html files change
./start-with-k8s.sh
```
