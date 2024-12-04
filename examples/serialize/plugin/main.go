package main

import (
	"log"

	"github.com/cvhariharan/plugin"
	"github.com/cvhariharan/plugin/example/serialize/plugin/shared"
)

// Name of this plugin, should uniquely identify this in the discovery service
// Can also be set from the env variable if multiple instances should be made available
const Name = "test"

func main() {
	p := &shared.TestPlugin{}
	log.Fatal(plugin.Serve(p, plugin.PluginServeOptions{Name: Name}))
}
