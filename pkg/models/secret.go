package models

// Secret represents a secret (without value)
type Secret struct {
	Name      string `json:"name"`
	VaultName string `json:"vault"`
	Provider  string `json:"provider"`
	Enabled   bool   `json:"enabled,omitempty"`
}

// SecretValue includes the actual secret value
type SecretValue struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	VaultName string `json:"vault"`
	Provider  string `json:"provider"`
}
