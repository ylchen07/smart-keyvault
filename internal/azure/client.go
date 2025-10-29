package azure

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

// Client executes Azure CLI commands
type Client struct {
	timeout time.Duration
}

// NewClient creates a new Azure CLI client
func NewClient() *Client {
	return &Client{
		timeout: 30 * time.Second,
	}
}

// Execute runs an az CLI command with the given arguments
func (c *Client) Execute(ctx context.Context, args ...string) ([]byte, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	// Build command
	cmd := exec.CommandContext(ctx, "az", args...)

	// Execute and capture output
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check if it's a context timeout
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("command timed out after %v", c.timeout)
		}

		// Check for common Azure CLI errors
		errMsg := string(output)
		if len(errMsg) > 0 {
			return nil, fmt.Errorf("azure cli error: %s", errMsg)
		}

		return nil, fmt.Errorf("azure cli command failed: %w", err)
	}

	return output, nil
}
