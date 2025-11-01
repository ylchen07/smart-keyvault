package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/ylchen07/smart-keyvault/internal/azure"
	"github.com/ylchen07/smart-keyvault/internal/clipboard"
	"github.com/ylchen07/smart-keyvault/internal/config"
	"github.com/ylchen07/smart-keyvault/internal/hashicorp"
	"github.com/ylchen07/smart-keyvault/internal/output"
	"github.com/ylchen07/smart-keyvault/internal/provider"
	"github.com/ylchen07/smart-keyvault/pkg/models"
)

var (
	providerName string
	instanceName string // New: instance name for multi-instance providers
	vaultName    string
	secretName   string
	formatType   string
	copyToClip   bool
	configPath   string // New: optional config file path

	// Global config loaded once
	appConfig *config.Config
)

func init() {
	// Register providers
	provider.Register("azure", azure.NewProvider)
	provider.Register("hashicorp", hashicorp.NewProvider)
}

// loadConfig loads the application config
func loadConfig() error {
	var err error

	if configPath != "" {
		appConfig, err = config.LoadFromFile(configPath)
	} else {
		appConfig, err = config.Load()
	}

	// If config file not found, create a minimal default config
	if err != nil {
		appConfig = &config.Config{
			Defaults: config.Defaults{},
			Providers: config.Providers{
				Azure:     &config.AzureConfig{Enabled: true, Instances: []config.AzureInstance{}},
				Hashicorp: &config.HashicorpConfig{Enabled: true, Instances: []config.HashicorpInstance{}},
			},
			FZF:     config.FZFConfig{Height: "40%", Border: "rounded", Preview: false},
			Filters: config.Filters{EnabledOnly: true},
		}
	}

	return nil
}

// getProviderConfig creates a provider.Config for the specified provider and instance
func getProviderConfig(providerName, instanceName string) (*provider.Config, error) {
	if appConfig == nil {
		if err := loadConfig(); err != nil {
			return nil, err
		}
	}

	cfg := &provider.Config{
		Name:     providerName,
		Settings: make(map[string]interface{}),
	}

	switch providerName {
	case "azure":
		var instance *config.AzureInstance
		var err error

		if instanceName != "" {
			instance, err = appConfig.GetAzureInstance(instanceName)
		} else {
			instance, err = appConfig.GetDefaultAzureInstance()
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get Azure instance: %w", err)
		}

		cfg.Settings["subscription_id"] = instance.SubscriptionID

	case "hashicorp":
		var instance *config.HashicorpInstance
		var err error

		if instanceName != "" {
			instance, err = appConfig.GetHashicorpInstance(instanceName)
		} else {
			instance, err = appConfig.GetDefaultHashicorpInstance()
		}

		if err != nil {
			return nil, fmt.Errorf("failed to get Hashicorp instance: %w", err)
		}

		cfg.Settings["address"] = instance.Address
		cfg.Settings["token"] = instance.Token
		cfg.Settings["namespace"] = instance.Namespace

	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}

	return cfg, nil
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "smart-keyvault",
		Short: "A multi-provider CLI for secret management",
		Long:  `Smart KeyVault provides a unified interface for browsing and retrieving secrets from Azure KeyVault, Hashicorp Vault, and more.`,
	}

	// Add commands
	rootCmd.AddCommand(listProvidersCmd())
	rootCmd.AddCommand(listVaultsCmd())
	rootCmd.AddCommand(listSecretsCmd())
	rootCmd.AddCommand(getSecretCmd())
	rootCmd.AddCommand(walkSecretsCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// listProvidersCmd returns the list-providers command
func listProvidersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-providers",
		Short: "List available secret providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			providers := provider.ListProviders()

			// Get formatter
			format := output.Format(formatType)
			formatter, err := output.GetFormatter(format)
			if err != nil {
				return err
			}

			// Format and output
			result, err := formatter.FormatProviders(providers)
			if err != nil {
				return err
			}

			fmt.Println(result)
			return nil
		},
	}

	cmd.Flags().StringVarP(&formatType, "format", "f", "plain", "Output format (plain, json)")
	return cmd
}

// listVaultsCmd returns the list-vaults command
func listVaultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-vaults",
		Short: "List all vaults from a provider",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get provider config
			cfg, err := getProviderConfig(providerName, instanceName)
			if err != nil {
				return err
			}

			// Get provider
			p, err := provider.GetProvider(providerName, cfg)
			if err != nil {
				return err
			}

			// List vaults
			ctx := context.Background()
			vaults, err := p.ListVaults(ctx)
			if err != nil {
				return err
			}

			// Get formatter
			format := output.Format(formatType)
			formatter, err := output.GetFormatter(format)
			if err != nil {
				return err
			}

			// Format and output
			result, err := formatter.FormatVaults(vaults)
			if err != nil {
				return err
			}

			fmt.Println(result)
			return nil
		},
	}

	cmd.Flags().StringVarP(&providerName, "provider", "p", "", "Provider name (azure, hashicorp)")
	cmd.Flags().StringVarP(&instanceName, "instance", "i", "", "Instance name (optional, uses default if not specified)")
	cmd.Flags().StringVarP(&formatType, "format", "f", "plain", "Output format (plain, json)")
	cmd.Flags().StringVar(&configPath, "config", "", "Config file path (optional)")
	cmd.MarkFlagRequired("provider")
	return cmd
}

// listSecretsCmd returns the list-secrets command
func listSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-secrets",
		Short: "List all secrets in a vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get provider config
			cfg, err := getProviderConfig(providerName, instanceName)
			if err != nil {
				return err
			}

			// Get provider
			p, err := provider.GetProvider(providerName, cfg)
			if err != nil {
				return err
			}

			// List secrets
			ctx := context.Background()
			secrets, err := p.ListSecrets(ctx, vaultName)
			if err != nil {
				return err
			}

			// Get formatter
			format := output.Format(formatType)
			formatter, err := output.GetFormatter(format)
			if err != nil {
				return err
			}

			// Format and output
			result, err := formatter.FormatSecrets(secrets)
			if err != nil {
				return err
			}

			fmt.Println(result)
			return nil
		},
	}

	cmd.Flags().StringVarP(&providerName, "provider", "p", "", "Provider name (azure, hashicorp)")
	cmd.Flags().StringVarP(&instanceName, "instance", "i", "", "Instance name (optional, uses default if not specified)")
	cmd.Flags().StringVarP(&vaultName, "vault", "v", "", "Vault name")
	cmd.Flags().StringVarP(&formatType, "format", "f", "plain", "Output format (plain, json)")
	cmd.Flags().StringVar(&configPath, "config", "", "Config file path (optional)")
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("vault")
	return cmd
}

// getSecretCmd returns the get-secret command
func getSecretCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-secret",
		Short: "Get a secret value",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get provider config
			cfg, err := getProviderConfig(providerName, instanceName)
			if err != nil {
				return err
			}

			// Get provider
			p, err := provider.GetProvider(providerName, cfg)
			if err != nil {
				return err
			}

			// Get secret
			ctx := context.Background()
			secret, err := p.GetSecret(ctx, vaultName, secretName)
			if err != nil {
				return err
			}

			// Copy to clipboard if requested
			if copyToClip {
				if err := clipboard.Copy(secret.Value); err != nil {
					return fmt.Errorf("failed to copy to clipboard: %w", err)
				}
				fmt.Fprintf(os.Stderr, "Secret '%s' copied to clipboard!\n", secretName)
			} else {
				// Output the secret value
				fmt.Println(secret.Value)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&providerName, "provider", "p", "", "Provider name (azure, hashicorp)")
	cmd.Flags().StringVarP(&instanceName, "instance", "i", "", "Instance name (optional, uses default if not specified)")
	cmd.Flags().StringVarP(&vaultName, "vault", "v", "", "Vault name")
	cmd.Flags().StringVarP(&secretName, "name", "n", "", "Secret name")
	cmd.Flags().BoolVarP(&copyToClip, "copy", "c", false, "Copy secret to clipboard")
	cmd.Flags().StringVar(&configPath, "config", "", "Config file path (optional)")
	cmd.MarkFlagRequired("provider")
	cmd.MarkFlagRequired("vault")
	cmd.MarkFlagRequired("name")
	return cmd
}

// walkSecretsCmd returns the walk-secrets command
func walkSecretsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "walk-secrets",
		Short: "Walk through all secrets in vaults and retrieve their values",
		Long:  `Walk through all accessible vaults (or a specific vault) and retrieve all secret values, outputting them in a structured format grouped by vault.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Load config
			if err := loadConfig(); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Get provider config
			cfg, err := getProviderConfig(providerName, instanceName)
			if err != nil {
				return err
			}

			// Get provider
			p, err := provider.GetProvider(providerName, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()

			// Determine which vaults to process
			var vaults []*models.Vault
			if vaultName != "" {
				// Single vault specified
				vaults = []*models.Vault{{Name: vaultName}}
			} else {
				// Get all vaults
				allVaults, err := p.ListVaults(ctx)
				if err != nil {
					return fmt.Errorf("failed to list vaults: %w", err)
				}
				vaults = allVaults
			}

			// Walk through each vault and collect all secrets with values
			secretsByVault := make(map[string][]*models.SecretValue)

			for _, vault := range vaults {
				// List secrets in vault
				secrets, err := p.ListSecrets(ctx, vault.Name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: failed to list secrets in vault %s: %v\n", vault.Name, err)
					continue
				}

				// Get value for each secret
				var secretValues []*models.SecretValue
				for _, secret := range secrets {
					secretValue, err := p.GetSecret(ctx, vault.Name, secret.Name)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to get secret %s in vault %s: %v\n", secret.Name, vault.Name, err)
						continue
					}
					secretValues = append(secretValues, secretValue)
				}

				secretsByVault[vault.Name] = secretValues
			}

			// Get formatter
			format := output.Format(formatType)
			formatter, err := output.GetFormatter(format)
			if err != nil {
				return err
			}

			// Format and output
			result, err := formatter.FormatWalkSecrets(secretsByVault)
			if err != nil {
				return err
			}

			fmt.Println(result)
			return nil
		},
	}

	cmd.Flags().StringVarP(&providerName, "provider", "p", "", "Provider name (azure, hashicorp)")
	cmd.Flags().StringVarP(&instanceName, "instance", "i", "", "Instance name (optional, uses default if not specified)")
	cmd.Flags().StringVarP(&vaultName, "vault", "v", "", "Vault name (optional - if not specified, walks all vaults)")
	cmd.Flags().StringVarP(&formatType, "format", "f", "json", "Output format (plain, json)")
	cmd.Flags().StringVar(&configPath, "config", "", "Config file path (optional)")
	cmd.MarkFlagRequired("provider")
	return cmd
}
