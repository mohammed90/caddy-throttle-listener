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

func TestLatency(t *testing.T) {
	latency, _ := caddy.ParseDuration("50ms")
	listener := Listener{
		Latency: caddy.Duration(latency),
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

		// Send a small message
		_, err = conn.Write([]byte("test"))
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
	}()

	// Wait for connection
	time.Sleep(10 * time.Millisecond)

	conn, err := wrappedLn.Accept()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	defer conn.Close()

	// Measure time for read operation (should include latency)
	start := time.Now()
	buf := make([]byte, 4)
	_, err = conn.Read(buf)
	duration := time.Since(start)

	if err != nil && err != io.EOF {
		t.Errorf("expected no error, got %s", err)
	}

	// Should take at least 50ms due to latency
	if duration < 40*time.Millisecond {
		t.Errorf("expected latency of at least 40ms, got %v", duration)
	}
}

func TestJitter(t *testing.T) {
	jitter, _ := caddy.ParseDuration("30ms")
	listener := Listener{
		Jitter: caddy.Duration(jitter),
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

	// Collect multiple timing samples to verify jitter variance
	var durations []time.Duration

	for i := 0; i < 5; i++ {
		go func() {
			conn, err := net.Dial("tcp", ln.Addr().String())
			if err != nil {
				t.Errorf("expected no error, got %s", err)
			}
			defer conn.Close()

			_, err = conn.Write([]byte("test"))
			if err != nil {
				t.Errorf("expected no error, got %s", err)
			}
		}()

		time.Sleep(10 * time.Millisecond)

		conn, err := wrappedLn.Accept()
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}

		start := time.Now()
		buf := make([]byte, 4)
		_, err = conn.Read(buf)
		duration := time.Since(start)
		conn.Close()

		if err != nil && err != io.EOF {
			t.Errorf("expected no error, got %s", err)
		}

		durations = append(durations, duration)
	}

	// Verify that we have some variance in timing due to jitter
	// At least one should be significantly different from another
	hasVariance := false
	for i := 0; i < len(durations)-1; i++ {
		diff := durations[i] - durations[i+1]
		if diff < 0 {
			diff = -diff
		}
		if diff > 5*time.Millisecond {
			hasVariance = true
			break
		}
	}

	if !hasVariance {
		t.Errorf("expected jitter to create timing variance, but all durations were similar: %v", durations)
	}
}

func TestLatencyAndJitter(t *testing.T) {
	latency, _ := caddy.ParseDuration("20ms")
	jitter, _ := caddy.ParseDuration("10ms")
	listener := Listener{
		Latency: caddy.Duration(latency),
		Jitter:  caddy.Duration(jitter),
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

		_, err = conn.Write([]byte("test"))
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
	}()

	time.Sleep(10 * time.Millisecond)

	conn, err := wrappedLn.Accept()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	defer conn.Close()

	start := time.Now()
	buf := make([]byte, 4)
	_, err = conn.Read(buf)
	duration := time.Since(start)

	if err != nil && err != io.EOF {
		t.Errorf("expected no error, got %s", err)
	}

	// Should take at least the base latency (20ms)
	if duration < 15*time.Millisecond {
		t.Errorf("expected delay of at least 15ms (base latency), got %v", duration)
	}

	// Should not exceed latency + jitter (30ms) by much
	if duration > 50*time.Millisecond {
		t.Errorf("expected delay of less than 50ms (latency + jitter + margin), got %v", duration)
	}
}

func TestThrottlingWithChaos(t *testing.T) {
	latency, _ := caddy.ParseDuration("10ms")
	jitter, _ := caddy.ParseDuration("5ms")
	listener := Listener{
		Up:      "1KiB",
		Down:    "1KiB",
		Latency: caddy.Duration(latency),
		Jitter:  caddy.Duration(jitter),
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

		data := make([]byte, 1024)
		_, err = conn.Write(data)
		if err != nil {
			t.Errorf("expected no error, got %s", err)
		}
	}()

	time.Sleep(10 * time.Millisecond)

	conn, err := wrappedLn.Accept()
	if err != nil {
		t.Errorf("expected no error, got %s", err)
	}
	defer conn.Close()

	// Measure combined effect of throttling + chaos
	start := time.Now()
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	duration := time.Since(start)

	if err != nil && err != io.EOF {
		t.Errorf("expected no error, got %s", err)
	}
	if n != 1024 {
		t.Errorf("expected to read 1024 bytes, got %d", n)
	}

	// Should have some delay from chaos features
	if duration < 5*time.Millisecond {
		t.Errorf("expected some delay from chaos features, got %v", duration)
	}
}
