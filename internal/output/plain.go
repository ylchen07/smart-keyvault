package output

import (
	"strings"

	"github.com/ylchen07/smart-keyvault/pkg/models"
)

// PlainFormatter outputs plain text (one item per line)
type PlainFormatter struct{}

// NewPlainFormatter creates a new plain text formatter
func NewPlainFormatter() *PlainFormatter {
	return &PlainFormatter{}
}

// FormatVaults formats vaults as plain text (one name per line)
func (f *PlainFormatter) FormatVaults(vaults []*models.Vault) (string, error) {
	if len(vaults) == 0 {
		return "", nil
	}

	names := make([]string, len(vaults))
	for i, v := range vaults {
		names[i] = v.Name
	}

	return strings.Join(names, "\n"), nil
}

// FormatSecrets formats secrets as plain text (one name per line)
func (f *PlainFormatter) FormatSecrets(secrets []*models.Secret) (string, error) {
	if len(secrets) == 0 {
		return "", nil
	}

	names := make([]string, len(secrets))
	for i, s := range secrets {
		names[i] = s.Name
	}

	return strings.Join(names, "\n"), nil
}

// FormatProviders formats provider names as plain text (one per line)
func (f *PlainFormatter) FormatProviders(providers []string) (string, error) {
	if len(providers) == 0 {
		return "", nil
	}

	return strings.Join(providers, "\n"), nil
}
