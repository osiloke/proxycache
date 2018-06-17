package proxy

// https://github.com/golang/go/issues/17393
import (
	"context"
	"net"
	"reflect"
	"syscall"
)

// SilenceSIGPIPE configures the net.Conn in a way that silences SIGPIPEs with
// the SO_NOSIGPIPE socket option.
func SilenceSIGPIPE(c net.Conn) error {
	// use reflection until https://github.com/golang/go/issues/9661 is fixed
	v := reflect.ValueOf(c).Elem().FieldByName("fd").Elem().FieldByName("sysfd")
	if !v.IsValid() {
		return nil
	}
	fd := int(v.Int())
	return syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_NOSIGPIPE, 1)
}

// NoSIGPIPEDialer returns a dialer that won't SIGPIPE should a connection
// actually SIGPIPE. This prevents the debugger from intercepting the signal
// even though this is normal behaviour.
type NoSIGPIPEDialer net.Dialer

func (d *NoSIGPIPEDialer) handle(c net.Conn, err error) (net.Conn, error) {
	if err != nil {
		return nil, err
	}
	if err := SilenceSIGPIPE(c); err != nil {
		c.Close()
		return nil, err
	}
	return c, err
}

func (d *NoSIGPIPEDialer) Dial(network, address string) (net.Conn, error) {
	c, err := (*net.Dialer)(d).Dial(network, address)
	return d.handle(c, err)
}

func (d *NoSIGPIPEDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	c, err := (*net.Dialer)(d).DialContext(ctx, network, address)
	return d.handle(c, err)
}
