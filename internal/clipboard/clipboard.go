package clipboard

import (
	"context"
	"fmt"

	"github.com/gopasspw/clipboard"
)

// Copy copies a password/secret to the system clipboard
// Uses WritePassword which may provide additional security features
func Copy(text string) error {
	ctx := context.Background()
	if err := clipboard.WritePassword(ctx, []byte(text)); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}
	return nil
}

// Read reads text from the system clipboard
func Read() (string, error) {
	ctx := context.Background()
	text, err := clipboard.ReadAllString(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read from clipboard: %w", err)
	}
	return text, nil
}

// IsAvailable checks if clipboard functionality is available
func IsAvailable() bool {
	// Check if clipboard is supported on this platform
	return !clipboard.IsUnsupported()
}
