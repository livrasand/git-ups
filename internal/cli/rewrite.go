package cli

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/livrasand/git-ups/internal/config"
)

func ServerBase() (string, error) {
	return config.GatewayURL()
}

var hostPrefix = map[string]string{
	"github.com":   "gh",
	"gitlab.com":   "gl",
	"codeberg.org": "cb",
}

func RewriteURL(base, raw string) (string, error) {
	base = strings.TrimRight(base, "/")
	u, err := parseRepoURL(raw)
	if err != nil {
		return "", err
	}

	prefix, ok := hostPrefix[u.Hostname()]
	if !ok {
		return "", fmt.Errorf("host not supported by gitGost: %s", u.Hostname())
	}

	parts := splitPath(u.Path)
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid repository URL: %s", raw)
	}
	owner, repo := parts[0], strings.TrimSuffix(parts[1], ".git")
	if owner == "" || repo == "" || !validSegment(owner) || !validSegment(repo) {
		return "", fmt.Errorf("invalid repository URL: %s", raw)
	}

	return fmt.Sprintf("%s/v1/%s/%s/%s", base, prefix, owner, repo), nil
}

func parseRepoURL(raw string) (*url.URL, error) {
	if strings.Contains(raw, "://") {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, fmt.Errorf("invalid URL: %w", err)
		}
		if u.Hostname() == "" {
			return nil, fmt.Errorf("invalid URL: %s", raw)
		}
		return u, nil
	}

	at := strings.LastIndex(raw, "@")
	colon := strings.Index(raw, ":")
	if at >= 0 && colon > at {
		host := raw[at+1 : colon]
		path := raw[colon+1:]
		return &url.URL{Scheme: "https", Host: host, Path: "/" + path}, nil
	}
	return nil, fmt.Errorf("invalid URL: %s", raw)
}

func splitPath(p string) []string {
	var out []string
	for _, part := range strings.Split(p, "/") {
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func validSegment(s string) bool {
	if len(s) == 0 || len(s) > 100 {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9',
			r == '-', r == '_', r == '.':
		default:
			return false
		}
	}
	return !strings.Contains(s, "..")
}
