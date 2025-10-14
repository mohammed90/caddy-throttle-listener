package caddythrottlelistener

import (
	"math/rand"
	"net"
	"strings"
	"time"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/conduitio/bwlimit"
	"github.com/dustin/go-humanize"
)

func init() {
	caddy.RegisterModule(Listener{})
}

// The `throttle` listener limits the bandwidth of the connection to the
// given values and can inject chaos engineering features like latency and jitter.
type Listener struct {
	// Up is the maximum upload speed. If not set, there is no limit.
	// The value is parsed using the `go-humanize` package and accepts
	// both the SI and the IEC prefixes. Values without units are interepreted
	// as bytes.
	Up string `json:"up,omitempty"`

	// Down is the maximum upload speed. If not set, there is no limit.
	// The value is parsed using the `go-humanize` package and accepts
	// both the SI and the IEC prefixes. Values without units are interepreted
	// as bytes.
	Down string `json:"down,omitempty"`

	// Latency adds a fixed delay to each read/write operation.
	// Accepts duration strings like "10ms", "1s", "500us".
	Latency caddy.Duration `json:"latency,omitempty"`

	// Jitter adds random delay (0 to jitter value) to each read/write operation.
	// Accepts duration strings like "5ms", "100ms", "1s".
	Jitter caddy.Duration `json:"jitter,omitempty"`

	up      bwlimit.Byte
	down    bwlimit.Byte
	latency time.Duration
	jitter  time.Duration
}

// CaddyModule returns the Caddy module information.
func (Listener) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "caddy.listeners.throttle",
		New: func() caddy.Module { return new(Listener) },
	}
}

// Provision implements caddy.Provisioner.
func (l *Listener) Provision(caddy.Context) error {
	l.up, l.down = 0, 0
	l.latency, l.jitter = 0, 0

	if up := strings.TrimSpace(l.Up); up != "" {
		u, err := humanize.ParseBytes(l.Up)
		if err != nil {
			return err
		}
		l.up = bwlimit.Byte(u)
	}
	if down := strings.TrimSpace(l.Down); down != "" {
		d, err := humanize.ParseBytes(l.Down)
		if err != nil {
			return err
		}
		l.down = bwlimit.Byte(d)
	}
	l.latency = time.Duration(l.Latency)
	l.jitter = time.Duration(l.Jitter)
	return nil
}

// WrapListener implements caddy.ListenerWrapper.
func (l *Listener) WrapListener(parent net.Listener) net.Listener {
	throttled := bwlimit.NewListener(parent, l.down, l.up)
	if l.latency > 0 || l.jitter > 0 {
		return &chaoticListener{
			Listener: throttled,
			latency:  l.latency,
			jitter:   l.jitter,
		}
	}
	return throttled
}

// chaoticListener wraps a listener to add chaos engineering features
type chaoticListener struct {
	net.Listener
	latency time.Duration
	jitter  time.Duration
}

func (cl *chaoticListener) Accept() (net.Conn, error) {
	conn, err := cl.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return &chaoticConn{
		Conn:    conn,
		latency: cl.latency,
		jitter:  cl.jitter,
	}, nil
}

// chaoticConn wraps a connection to inject latency and jitter
type chaoticConn struct {
	net.Conn
	latency time.Duration
	jitter  time.Duration
}

func (cc *chaoticConn) Read(b []byte) (int, error) {
	cc.addDelay()
	return cc.Conn.Read(b)
}

func (cc *chaoticConn) Write(b []byte) (int, error) {
	cc.addDelay()
	return cc.Conn.Write(b)
}

func (cc *chaoticConn) addDelay() {
	totalDelay := cc.latency
	if cc.jitter > 0 {
		jitterDelay := time.Duration(rand.Int63n(int64(cc.jitter)))
		totalDelay += jitterDelay
	}
	if totalDelay > 0 {
		time.Sleep(totalDelay)
	}
}

// UnmarshalCaddyfile implements caddyfile.Unmarshaler.
func (l *Listener) UnmarshalCaddyfile(d *caddyfile.Dispenser) error {
	d.Next()
	for nesting := d.Nesting(); d.NextBlock(nesting); {
		switch d.Val() {
		case "up":
			if !d.NextArg() {
				return d.ArgErr()
			}
			l.Up = d.Val()
		case "down":
			if !d.NextArg() {
				return d.ArgErr()
			}
			l.Down = d.Val()
		case "latency":
			if !d.NextArg() {
				return d.ArgErr()
			}
			latency, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return d.Errf("invalid latency duration %s: %v", d.Val(), err)
			}
			l.Latency = caddy.Duration(latency)
		case "jitter":
			if !d.NextArg() {
				return d.ArgErr()
			}
			jitter, err := caddy.ParseDuration(d.Val())
			if err != nil {
				return d.Errf("invalid jitter duration %s: %v", d.Val(), err)
			}
			l.Jitter = caddy.Duration(jitter)
		default:
			return d.Errf("unrecognized subdirective %s", d.Val())
		}
	}
	return nil
}

var (
	_ caddy.Module          = (*Listener)(nil)
	_ caddy.Provisioner     = (*Listener)(nil)
	_ caddy.ListenerWrapper = (*Listener)(nil)
	_ caddyfile.Unmarshaler = (*Listener)(nil)
)
