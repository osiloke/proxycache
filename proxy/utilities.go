package proxy

import (
	"fmt"
	"regexp"
)

var re = regexp.MustCompile(`((?:(?:https?|ftp|file)))`)

// ReplaceHLSUrls replace hls urls
func ReplaceHLSUrls(hlsRaw []byte, proxyServerURL string) ([]byte, error) {
	s := re.ReplaceAllString(string(hlsRaw), fmt.Sprintf(`%s$1`, proxyServerURL))
	return []byte(s), nil
}
