package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
)

// HTTPConnectProxyServer is a simple HTTP CONNECT proxy.
type HTTPConnectProxyServer struct {
	Address string
	server http.Server

	waitForFirstClientPacket bool
}

// StartHTTPConnectProxyServer starts up an [http.Server] for HTTP CONNECT proxy
// handling on a random port localhost.
func StartHTTPConnectProxyServer(port int, waitForFirstClientPacket bool) (*HTTPConnectProxyServer, error) {
	l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		return nil, err
	}

	srv := &HTTPConnectProxyServer{Address: l.Addr().String(), waitForFirstClientPacket: waitForFirstClientPacket}
	srv.server.Handler = http.HandlerFunc(srv.handler)
	go func() {
		if err := srv.server.Serve(l); err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP CONNECT proxy failed: %v\n", err)
		}
	}()
	return srv, nil
}

func (h *HTTPConnectProxyServer) Close() error {
	return h.server.Close()
}

func (h *HTTPConnectProxyServer) handler(w http.ResponseWriter, r *http.Request) {
	// Much of this taken from TestTransportProxy in Go source
	// https://github.com/golang/go/blob/805f6b3f5db714ce8f7dae2776748f6df96f288b/src/net/http/transport_test.go#L1406-L1441
	if r.Method != "CONNECT" {
		http.Error(w, "CONNECT only", http.StatusMethodNotAllowed)
		return
	}
	fmt.Printf("Got HTTP proxy request; host=%s\n", r.Host)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		panic("no hijack iface")
	}
	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		panic("hijack failed")
	}

	targetConn, err := net.Dial("tcp", r.URL.Host)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upstream conn failed: %v", err), http.StatusBadGateway)
		return
	}

	res := &http.Response{
		StatusCode: http.StatusOK,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
	}

	if err := res.Write(clientConn); err != nil {
		panic(fmt.Sprintf("Writing 200 OK failed: %v", err))
	}

	if (h.waitForFirstClientPacket) {
		// Artifically wait for the first client packet before falling into full duplex mode.
		// This is a workaround for the issue being demonstrated here.
		buf := make([]byte, 32 * 1024)
		readBytes, err := clientConn.Read(buf)
		if err != nil {
			panic(fmt.Sprintf("Expected client to send a first packet: %v", err))
		}
		if _, err := targetConn.Write(buf[:readBytes]); err != nil {
			panic(fmt.Sprintf("Sending client's first packet to server failed: %v", err))
		}
	}

	go io.Copy(targetConn, clientConn)
	go func() {
		io.Copy(clientConn, targetConn)
		targetConn.Close()
	}()
}

func main() {
	args := os.Args[1:]
	port := 8888
	waitForFirstClientPacket := false
	if (len(args) >= 2 && args[0] == "--port") {
		port, _ = strconv.Atoi(args[1])
		args = args[2:]
	}
	if (len(args) >= 1 && args[0] == "--wait-for-first-client-packet") {
		waitForFirstClientPacket = true
		args = args[1:]
	}
	if len(args) > 0 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--port PORT]\n", os.Args[0])
		os.Exit(1)
	}

	proxyServer, err := StartHTTPConnectProxyServer(port, waitForFirstClientPacket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not start proxy server: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Started HTTP CONNECT proxy server on http://%s\n", proxyServer.Address)
	defer proxyServer.Close()

    // Sleep until the process is interrupted
	sigs := make(chan os.Signal, 1)
    signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
    <-sigs
	fmt.Printf("Press Ctrl+C to exit\n")
}