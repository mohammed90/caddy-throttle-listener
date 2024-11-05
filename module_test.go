package caddythrottlelistener

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/caddyserver/caddy/v2"
)

func TestThrottling(t *testing.T) {
	listener := Listener{
		Up:   "1KiB",
		Down: "1KiB",
	}
	ctx := caddy.Context{}
	err := listener.Provision(ctx)
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	defer ln.Close()

	wrappedLn := listener.WrapListener(ln)

	go func() {
		conn, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
		defer conn.Close()

		// Send 10KB of data to the server
		data := make([]byte, 10*1024)
		_, err = conn.Write(data)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
	}()

	// Wait for the connection to be established
	time.Sleep(100 * time.Millisecond)

	// Accept the connection
	conn, err := wrappedLn.Accept()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	defer conn.Close()

	// Read 1KB of data from the connection
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("expected no error, got %s", err)
	}
	if n != 1024 {
		t.Errorf("expected to read 1024 bytes, got %d", n)
	}

	// Wait for the throttling to kick in
	time.Sleep(100 * time.Millisecond)

	// Read the remaining 9KB of data from the connection
	buf = make([]byte, 9*1024)
	n, err = conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("expected no error, got %s", err)
	}
	if n != 1024 {
		t.Errorf("expected to read 1024 bytes, got %d", n)
	}
}

func TestThrottlingUpDown(t *testing.T) {
	listener := Listener{
		Up:   "1KiB",
		Down: "2KiB",
	}
	ctx := caddy.Context{}
	err := listener.Provision(ctx)
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}

	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	defer ln.Close()

	wrappedLn := listener.WrapListener(ln)

	go func() {
		conn, err := net.Dial("tcp", ln.Addr().String())
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
		defer conn.Close()

		// Send 2KB of data to the server
		data := make([]byte, 2*1024)
		_, err = conn.Write(data)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
	}()

	// Wait for the connection to be established
	time.Sleep(100 * time.Millisecond)

	// Accept the connection
	conn, err := wrappedLn.Accept()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	defer conn.Close()

	// Read 1KB of data from the connection
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		t.Errorf("expected no error, got %s", err)
	}
	if n != 1024 {
		t.Errorf("expected to read 1024 bytes, got %d", n)
	}

	// Send 1KB of data back to the client
	_, err = conn.Write(buf)
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
}
