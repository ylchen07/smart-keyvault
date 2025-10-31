package output

import (
	"encoding/json"

	"github.com/ylchen07/smart-keyvault/pkg/models"
)

// JSONFormatter outputs JSON format
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// FormatVaults formats vaults as JSON
func (f *JSONFormatter) FormatVaults(vaults []*models.Vault) (string, error) {
	data, err := json.MarshalIndent(vaults, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatSecrets formats secrets as JSON
func (f *JSONFormatter) FormatSecrets(secrets []*models.Secret) (string, error) {
	data, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatProviders formats provider names as JSON
func (f *JSONFormatter) FormatProviders(providers []string) (string, error) {
	data, err := json.MarshalIndent(providers, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatWalkSecrets formats all secrets grouped by vault as JSON
func (f *JSONFormatter) FormatWalkSecrets(secretsByVault map[string][]*models.SecretValue) (string, error) {
	data, err := json.MarshalIndent(secretsByVault, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
