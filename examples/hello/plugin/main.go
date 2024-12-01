package main

import (
	"log"

	"github.com/cvhariharan/plugin"
	"github.com/cvhariharan/plugin/example/hello/plugin/shared"
)

// Name of this plugin, should uniquely identify this in the discovery service
// Can also be set from the env variable if multiple instances should be made available
const Name = "hello"

// This is the actual implementation
type HelloImpl struct{}

func (hi *HelloImpl) Greet() string {
	return "Hello World!"
}

func main() {
	p := &shared.HelloPlugin{Impl: &HelloImpl{}}
	log.Fatal(plugin.Serve(p, plugin.PluginServeOptions{Name: Name}))
}
