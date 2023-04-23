package main

import (
	"github.com/jreyesr/steampipe-plugin-blockchain/blockchain"
	"github.com/turbot/steampipe-plugin-sdk/v5/plugin"
)

func main() {
	blockchain.InitMetrics()

	plugin.Serve(&plugin.ServeOpts{
		PluginFunc: blockchain.Plugin})
}
