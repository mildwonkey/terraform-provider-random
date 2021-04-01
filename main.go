package main

import (
	plugin "github.com/hashicorp/go-plugin"
	tf6server "github.com/hashicorp/terraform-plugin-go/tfprotov6/server"
	random "github.com/mildwonkey/terraform-provider-random/internal/provider"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion: 6,
	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		GRPCServer:      plugin.DefaultGRPCServer,
		Plugins: plugin.PluginSet{
			"provider": &tf6server.GRPCProviderPlugin{
				GRPCProvider: random.Server,
			},
		},
	})
}
