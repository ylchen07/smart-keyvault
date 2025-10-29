package models

// ProviderInfo holds metadata about a provider
type ProviderInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Available   bool     `json:"available"`
	Features    []string `json:"features,omitempty"`
}
