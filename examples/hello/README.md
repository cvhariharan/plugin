## Hello Plugin
This example shows how to build a plugin and use it. The plugin can be found in `plugin` directory. The `shared` directory contains interfaces that will be used by both client and server.

### Build
First build the plugin
```bash
cd plugin
go build -o hello
```

Now you can run `main.go` under `example/hello` directory. This will launch the binary `hello` in a separate process and initialize a client to interact with the plugin. 