package blockchain

import (
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin/schema"
)

// This uses go-cty: https://github.com/zclconf/go-cty/blob/main/docs/gocty.md#converting-to-and-from-structs
type BlockchainConfig struct {
	// At least a parameter is required, otherwise the parsing code gets angry
	Placeholder *bool `cty:"placeholder"`
}

var ConfigSchema = map[string]*schema.Attribute{
	"placeholder": {Type: schema.TypeBool},
}

func ConfigInstance() interface{} {
	return &BlockchainConfig{}
}

// GetConfig :: retrieve and cast connection config from query data
func GetConfig(connection *plugin.Connection) BlockchainConfig {
	if connection == nil || connection.Config == nil {
		return BlockchainConfig{}
	}
	config, _ := connection.Config.(BlockchainConfig)
	return config
}

func (c BlockchainConfig) String() string {
	// There is no need to print the placeholder param, it does nothing
	return "BlockchainConfig{}"
}
