package commons

import (
	"errors"
	"net/url"
	"strings"
)

// CleanClientRoute is to verify if a path passed from client is a valid endpoint and removes query strings or other
// relatively-useless parameters to avoid flooding
func CleanClientRoute(route string) (string, error) {
	r, err := url.ParseRequestURI(route)
	if err != nil {
		return "", err
	}
	// the route (or a path) shall be absolute and doesn't have the Host and Scheme set
	if len(r.Path) > 128 || !strings.HasPrefix(r.Path, "/") {
		return "(malformed)", errors.New("malformed route")
	}
	return r.Path, nil
}
