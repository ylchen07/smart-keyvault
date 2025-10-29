package output

import (
	"github.com/ylchen07/smart-keyvault/pkg/models"
)

// Format represents the output format type
type Format string

const (
	// FormatPlain is plain text format (one item per line)
	FormatPlain Format = "plain"
	// FormatJSON is JSON format
	FormatJSON Format = "json"
)

// Formatter formats data for output
type Formatter interface {
	FormatVaults(vaults []*models.Vault) (string, error)
	FormatSecrets(secrets []*models.Secret) (string, error)
	FormatProviders(providers []string) (string, error)
}
