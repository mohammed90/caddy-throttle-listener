package caddythrottlelistener

import (
	"net"
	"strings"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/conduitio/bwlimit"
	"github.com/dustin/go-humanize"
)

func init() {
	caddy.RegisterModule(Listener{})
}

// The `throttle` listener limits the bandwidth of the connection to the
// given values.
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

	up   bwlimit.Byte
	down bwlimit.Byte
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
	return nil
}

// WrapListener implements caddy.ListenerWrapper.
func (l *Listener) WrapListener(parent net.Listener) net.Listener {
	return bwlimit.NewListener(parent, l.down, l.up)
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
