package validate

import (
	"net"
	"net/url"
	"regexp"
	"strings"
)

var domainRE = regexp.MustCompile(`^(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?)(?:\.(?i:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?))*$`)

func Domain(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || net.ParseIP(value) != nil {
		return false
	}
	return domainRE.MatchString(value)
}

func URL(value string) bool {
	u, err := url.ParseRequestURI(strings.TrimSpace(value))
	if err != nil {
		return false
	}
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}
