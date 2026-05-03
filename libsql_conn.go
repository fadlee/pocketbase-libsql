package main

import (
	"net/url"
	"strings"
)

func buildLibSQLConnectionString(rawURL string, token string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	if u.Scheme == "" || u.Host == "" {
		return "", errMissingSchemeOrHost{}
	}
	if token == "" {
		return rawURL, nil
	}

	q := u.Query()
	q.Set("authToken", token)
	u.RawQuery = q.Encode()

	return u.String(), nil
}

func maskLibSQLURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "<invalid-url>"
	}

	if u.RawQuery == "" {
		return u.String()
	}

	parts := strings.Split(u.RawQuery, "&")
	for i, part := range parts {
		key, _, _ := strings.Cut(part, "=")
		decodedKey, err := url.QueryUnescape(key)
		if err != nil {
			return "<invalid-url>"
		}
		if decodedKey == "authToken" {
			parts[i] = key + "=***"
		}
	}
	u.RawQuery = strings.Join(parts, "&")

	return u.String()
}

type errMissingSchemeOrHost struct{}

func (errMissingSchemeOrHost) Error() string {
	return "missing URL scheme or host"
}
