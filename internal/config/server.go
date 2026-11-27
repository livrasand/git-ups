package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

// DefaultGatewayURL is the gitGost gateway used when GITGOST_SERVER is not set.
const DefaultGatewayURL = "https://gitgost.livrasand.com"

// GatewayURL returns the base URL of the gitGost gateway. It uses the
// GITGOST_SERVER environment variable when set, and DefaultGatewayURL otherwise.
// A trailing slash is removed. Only http and https URLs are accepted.
func GatewayURL() (string, error) {
	v := strings.TrimSpace(os.Getenv("GITGOST_SERVER"))
	if v == "" {
		return DefaultGatewayURL, nil
	}
	v = strings.TrimRight(v, "/")
	u, err := url.Parse(v)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("invalid GITGOST_SERVER %q: must be an http:// or https:// URL", v)
	}
	return v, nil
}
