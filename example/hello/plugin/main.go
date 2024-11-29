package main

import (
	"github.com/cvhariharan/plugin"
	"github.com/cvhariharan/plugin/example/hello/plugin/shared"
)

// This is the actual implementation
type HelloImpl struct{}

func (hi *HelloImpl) Greet() string {
	return "Hello World!"
}

func main() {
	p := &shared.HelloPlugin{Impl: &HelloImpl{}}
	plugin.Serve(p)
}
