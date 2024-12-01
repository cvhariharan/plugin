package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/cvhariharan/plugin/catalog"
	"github.com/cvhariharan/plugin/store"
	"google.golang.org/grpc"
)

func main() {
	cs := store.NewMemCatalogStore()
	var opt []grpc.ServerOption
	srv := grpc.NewServer(opt...)
	catalog.NewCatalogServer(cs, srv)

	catalogListener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := srv.Serve(catalogListener); err != nil {
			log.Fatal(err)
		}
		wg.Done()
	}()

	// c, err := plugin.Load(
	// 	plugin.PluginLoadOptions{
	// 		Name: "hello",
	// 		Path: "plugin/hello",
	// 		PluginMap: map[string]plugin.Plugin{
	// 			"hello": &hello.HelloPlugin{},
	// 		},
	// 	},
	// 	cs,
	// )
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// client := c.(hello.Hello)

	// // use the client just like any normal object
	// fmt.Println(client.Greet())
	time.Sleep(10 * time.Second)
	fmt.Println(cs.Get("hello"))

	wg.Wait()
}
