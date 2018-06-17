package proxy

import (
	"fmt"
	"regexp"
)

var re = regexp.MustCompile(`((?:(?:https?|ftp|file):\/\/|www\.|ftp\.)(?:\([-A-Z0-9+&@#\/%=~_|$?!:,.]*\)|[-A-Z0-9+&@#\/%=~_|$?!:,.])*(?:\([-A-Z0-9+&@#\/%=~_|$?!:,.]*\)|[A-Z0-9+&@#\/%=~_|$]))`)

func replaceHLSUrls(hlsRaw []byte, proxyServerURL string) ([]byte, error) {
	s := re.ReplaceAllString(string(hlsRaw), fmt.Sprintf(`%s$1`, proxyServerURL))
	return []byte(s), nil
}
