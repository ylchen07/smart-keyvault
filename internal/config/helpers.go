package config

import (
	"fmt"
)

// GetAzureInstance returns an Azure instance by name
func (c *Config) GetAzureInstance(name string) (*AzureInstance, error) {
	if c.Providers.Azure == nil {
		return nil, fmt.Errorf("azure provider not configured")
	}

	for _, inst := range c.Providers.Azure.Instances {
		if inst.Name == name {
			return &inst, nil
		}
	}

	return nil, fmt.Errorf("azure instance '%s' not found", name)
}

// GetDefaultAzureInstance returns the default Azure instance
func (c *Config) GetDefaultAzureInstance() (*AzureInstance, error) {
	if c.Providers.Azure == nil {
		return nil, fmt.Errorf("azure provider not configured")
	}

	// Look for instance marked as default
	for _, inst := range c.Providers.Azure.Instances {
		if inst.Default {
			return &inst, nil
		}
	}

	// If no default, return first instance
	if len(c.Providers.Azure.Instances) > 0 {
		return &c.Providers.Azure.Instances[0], nil
	}

	return nil, fmt.Errorf("no azure instances configured")
}

// ListAzureInstances returns all Azure instances
func (c *Config) ListAzureInstances() []AzureInstance {
	if c.Providers.Azure == nil {
		return []AzureInstance{}
	}
	return c.Providers.Azure.Instances
}

// GetHashicorpInstance returns a Hashicorp Vault instance by name
func (c *Config) GetHashicorpInstance(name string) (*HashicorpInstance, error) {
	if c.Providers.Hashicorp == nil {
		return nil, fmt.Errorf("hashicorp provider not configured")
	}

	for _, inst := range c.Providers.Hashicorp.Instances {
		if inst.Name == name {
			return &inst, nil
		}
	}

	return nil, fmt.Errorf("hashicorp instance '%s' not found", name)
}

// GetDefaultHashicorpInstance returns the default Hashicorp Vault instance
func (c *Config) GetDefaultHashicorpInstance() (*HashicorpInstance, error) {
	if c.Providers.Hashicorp == nil {
		return nil, fmt.Errorf("hashicorp provider not configured")
	}

	// Look for instance marked as default
	for _, inst := range c.Providers.Hashicorp.Instances {
		if inst.Default {
			return &inst, nil
		}
	}

	// If no default, return first instance
	if len(c.Providers.Hashicorp.Instances) > 0 {
		return &c.Providers.Hashicorp.Instances[0], nil
	}

	return nil, fmt.Errorf("no hashicorp instances configured")
}

// ListHashicorpInstances returns all Hashicorp Vault instances
func (c *Config) ListHashicorpInstances() []HashicorpInstance {
	if c.Providers.Hashicorp == nil {
		return []HashicorpInstance{}
	}
	return c.Providers.Hashicorp.Instances
}

// IsProviderEnabled checks if a provider is enabled
func (c *Config) IsProviderEnabled(providerName string) bool {
	switch providerName {
	case "azure":
		return c.Providers.Azure != nil && c.Providers.Azure.Enabled
	case "hashicorp":
		return c.Providers.Hashicorp != nil && c.Providers.Hashicorp.Enabled
	default:
		return false
	}
}

// GetEnabledProviders returns a list of enabled provider names
func (c *Config) GetEnabledProviders() []string {
	var providers []string

	if c.Providers.Azure != nil && c.Providers.Azure.Enabled {
		providers = append(providers, "azure")
	}

	if c.Providers.Hashicorp != nil && c.Providers.Hashicorp.Enabled {
		providers = append(providers, "hashicorp")
	}

	return providers
}
