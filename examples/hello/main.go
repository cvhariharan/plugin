package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/cvhariharan/plugin"
	"github.com/cvhariharan/plugin/catalog"
	hello "github.com/cvhariharan/plugin/example/hello/plugin/shared"
	"github.com/cvhariharan/plugin/store"
)

func main() {

	// Start the discovery server to catalog remote plugins
	cs := store.NewMemCatalogStore()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := catalog.Serve(cs, ":50051"); err != nil {
			log.Println(err)
		}
		wg.Done()
	}()

	c, err := plugin.Load(
		plugin.PluginLoadOptions{
			Name: "hello",
			// If address is set, remote plugin will be loaded
			// Address: svc.Address,
			Path: "plugin/hello",
			PluginMap: map[string]plugin.Plugin{
				"hello": &hello.HelloPlugin{},
			},
		},
		cs,
	)
	if err != nil {
		log.Fatal(err)
	}

	client := c.(hello.Hello)

	// use the client just like any normal object
	fmt.Println(client.Greet())

	wg.Wait()
}
