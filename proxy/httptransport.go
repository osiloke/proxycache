package proxy

// https://github.com/golang/go/issues/17393
import (
	"net/http"
	"time"
)

// NoSIGPIPETransport is a default HTTP transport (configured in the same manner)
var NoSIGPIPETransport http.RoundTripper = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&NoSIGPIPEDialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func init() {
	// Install NoSIGPIPETransport as the default HTTP transport
	// http.DefaultTransport = NoSIGPIPETransport
	// Install NoSIGPIPEDialer as the default HTTP dialer
	defaultTransport := http.DefaultTransport.(*http.Transport)
	defaultTransport.DialContext = (&NoSIGPIPEDialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext
}
