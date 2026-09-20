package github

import "net/http"

func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", defaultApiVersion)

	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}
}
