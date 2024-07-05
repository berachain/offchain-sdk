package eth

import "net/url"

// trimProtocolAndPort removes the protocol (scheme) and port from a URL.
func trimProtocolAndPort(rawURL string) string {
	// Parse the URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	// Clear the protocol (scheme) and port
	parsedURL.Scheme = ""
	parsedURL.Host = parsedURL.Hostname() // Removes the port

	// Rebuild the URL without the protocol and port
	// We use RequestURI() to avoid the leading "//" when scheme is empty
	return parsedURL.RequestURI()
}

// formatTags converts a map of tags into format "key:value".
func formatTags(tags map[string]string) []string {
	if len(tags) == 0 {
		return nil
	}

	formattedTags := make([]string, 0, len(tags))
	for key, value := range tags {
		formattedTags = append(formattedTags, key+":"+value)
	}
	return formattedTags
}
