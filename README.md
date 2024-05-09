## To reproduce:

1. Install npm dependencies.

```console
$ npm i
```

2. Start the gRPC server and proxy in background; alternatively, you may run these in a distinct tabs.

```console
$ node src/greeter_server.js --port 6233 &
$ go run go/httpproxy.go --port 8888 &
```

3. Run the client 1000 times, without proxy

```console
$ node src/greeter_client.js --target 127.0.0.1:6233 --repeat 1000
Completed without error
```

4. Run again, this time with proxy

```console
$ export grpc_proxy=http://127.0.0.1:8888
$ node src/greeter_client.js --target 127.0.0.1:6233 --repeat 1000
Error: 13 INTERNAL: Received RST_STREAM with code 2 triggered by internal client error: Protocol error
(**redacted for brievety; see full error details in linked gist**)
{
  code: 13,
  details: 'Received RST_STREAM with code 2 triggered by internal client error: Protocol error',
  metadata: Metadata { internalRepr: Map(0) {}, options: {} }
}
Failed at iteration 43
```

5. Restart the proxy, this time with the "wait for client's first packet" patch

```console
$ killall  httpproxy
$ go run go/httpproxy.go --port 8888 --wait-for-first-client-packet &
```

6. Run the client 1000 times, with the patched proxy

```console
$ export grpc_proxy=http://127.0.0.1:8888
$ node src/greeter_client.js --target 127.0.0.1:6233 --repeat 1000
```
