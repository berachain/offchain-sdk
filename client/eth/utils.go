package eth

import "net/url"

// trimProtocolAndPort removes the protocol (scheme) and port from a URL.
func trimProtocolAndPort(rawURL string) string {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsedURL.Hostname() + parsedURL.Path
}
