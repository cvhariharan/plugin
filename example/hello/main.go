package main

import (
	"fmt"
	"log"

	"github.com/cvhariharan/plugin"
	hello "github.com/cvhariharan/plugin/example/hello/plugin/shared"
)

func main() {
	c, err := plugin.Load(
		plugin.PluginLoadOptions{
			Name: "hello",
			Path: "plugin/hello",
			PluginMap: map[string]plugin.Plugin{
				"hello": &hello.HelloPlugin{},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	client := c.(hello.Hello)
	fmt.Println(client.Greet())
}
