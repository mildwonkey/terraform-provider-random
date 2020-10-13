package main

import (
	plugin "github.com/hashicorp/go-plugin"
	tfprotov5server "github.com/hashicorp/terraform-plugin-go/tfprotov5/server"
	random "github.com/mildwonkey/terraform-provider-random/internal/provider"
)

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion: 5,
	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		GRPCServer:      plugin.DefaultGRPCServer,
		Plugins: plugin.PluginSet{
			"provider": &tfprotov5server.GRPCProviderPlugin{
				GRPCProvider: random.Server,
			},
		},
	})
}
