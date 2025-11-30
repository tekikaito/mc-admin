package rcon

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorcon/rcon"
)

type CommandExecutor interface {
	ExecuteCommand(cmd string) (string, error)
}

type MinecraftRconClient struct {
	Host     string
	Port     string
	Password string
	conn     *rcon.Conn
	mu       sync.Mutex
}

func NewMinecraftRconClient(host, port, password string) *MinecraftRconClient {
	return &MinecraftRconClient{
		Host:     host,
		Port:     port,
		Password: password,
	}
}

func (c *MinecraftRconClient) getConnectionString() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// connect establishes a new RCON connection
func (c *MinecraftRconClient) connect() error {
	connectionString := c.getConnectionString()
	conn, err := rcon.Dial(connectionString, c.Password)
	if err != nil {
		return fmt.Errorf("failed to connect to RCON server at %s: %w", connectionString, err)
	}
	c.conn = conn
	return nil
}

// disconnect closes the current RCON connection
func (c *MinecraftRconClient) disconnect() {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
}

// isConnected checks if the connection is still alive
func (c *MinecraftRconClient) isConnected() bool {
	if c.conn == nil {
		return false
	}
	// Try a simple command to verify connection is alive
	_, err := c.conn.Execute("list")
	return err == nil
}

// ensureConnected ensures a valid connection exists, reconnecting if necessary
func (c *MinecraftRconClient) ensureConnected() error {
	if c.isConnected() {
		return nil
	}

	c.disconnect()

	// Retry connection with exponential backoff
	maxRetries := 3
	for i := range maxRetries {
		err := c.connect()
		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			backoff := time.Duration(i+1) * time.Second
			time.Sleep(backoff)
		}
	}

	return fmt.Errorf("failed to establish RCON connection after %d attempts", maxRetries)
}

// ExecuteCommand executes an RCON command with automatic reconnection
func (c *MinecraftRconClient) ExecuteCommand(command string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return "", err
	}

	response, err := c.conn.Execute(command)
	if err != nil {
		// Try to reconnect once if command fails
		c.disconnect()
		if reconnectErr := c.connect(); reconnectErr != nil {
			return "", fmt.Errorf("command failed and reconnection failed: %w", err)
		}

		// Retry the command once
		response, err = c.conn.Execute(command)
		if err != nil {
			return "", fmt.Errorf("command failed after reconnection: %w", err)
		}
	}

	return response, nil
}

// Close closes the RCON connection gracefully
func (c *MinecraftRconClient) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.disconnect()
}
