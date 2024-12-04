package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/cvhariharan/plugin"
	"github.com/cvhariharan/plugin/catalog"
	test "github.com/cvhariharan/plugin/example/serialize/plugin/shared"
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
			Name: "test",
			// If address is set, remote plugin will be loaded
			// Address: svc.Address,
			Path: "plugin/test",
			// This specifies the exact plugin type
			Plugin: &test.TestPlugin{},
		},
		cs,
	)
	if err != nil {
		log.Fatal(err)
	}

	// cast to the interface we expect
	client := c.(test.Test)

	t := &test.TestObj{Data: "some test data"}
	// use the client just like any normal object
	fmt.Println(client.TestCall(t))

	wg.Wait()
}
