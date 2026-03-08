package infrastructure

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	researchapp "github.com/alty-cli/alty/internal/research/application"
	researchdomain "github.com/alty-cli/alty/internal/research/domain"
)

// DuckDuckGoSearchAdapter implements WebSearch using DuckDuckGo HTML search.
// Since no Go library exists for DuckDuckGo, this uses HTTP requests to the
// DuckDuckGo HTML endpoint and parses results.
type DuckDuckGoSearchAdapter struct {
	httpClient *http.Client
}

// Compile-time interface check.
var _ researchapp.WebSearch = (*DuckDuckGoSearchAdapter)(nil)

// NewDuckDuckGoSearchAdapter creates a DuckDuckGoSearchAdapter.
func NewDuckDuckGoSearchAdapter() *DuckDuckGoSearchAdapter {
	return &DuckDuckGoSearchAdapter{
		httpClient: &http.Client{},
	}
}

// Search executes a DuckDuckGo search query.
// Returns empty slice on any error (graceful degradation).
func (d *DuckDuckGoSearchAdapter) Search(
	ctx context.Context,
	query string,
	maxResults int,
) ([]researchdomain.WebSearchResult, error) {
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		slog.Warn("DuckDuckGo search: failed to create request", "error", err)
		return nil, nil
	}
	req.Header.Set("User-Agent", "alty-cli/1.0")

	resp, err := d.httpClient.Do(req)
	if err != nil {
		slog.Warn("DuckDuckGo search failed", "query", query, "error", err)
		return nil, nil
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Warn("DuckDuckGo search: failed to read response", "error", err)
		return nil, nil
	}

	return parseHTMLResults(string(body), maxResults), nil
}

// parseHTMLResults extracts search results from DuckDuckGo HTML response.
// This is a simplified parser that looks for result links and snippets.
func parseHTMLResults(html string, maxResults int) []researchdomain.WebSearchResult {
	var results []researchdomain.WebSearchResult

	// Look for result blocks in the HTML
	// DuckDuckGo HTML results contain class="result__a" for links
	// and class="result__snippet" for snippets
	blocks := strings.Split(html, "class=\"result__a\"")
	for i, block := range blocks {
		if i == 0 || len(results) >= maxResults {
			continue
		}

		href := extractAttribute(block, "href=\"")
		title := extractTextContent(block)
		snippet := ""
		if snipIdx := strings.Index(block, "class=\"result__snippet\""); snipIdx != -1 {
			snippet = extractTextContent(block[snipIdx:])
		}

		if href != "" && title != "" {
			results = append(results, researchdomain.NewWebSearchResult(href, title, snippet))
		}
	}

	return results
}

func extractAttribute(s, prefix string) string {
	idx := strings.Index(s, prefix)
	if idx == -1 {
		return ""
	}
	start := idx + len(prefix)
	end := strings.IndexByte(s[start:], '"')
	if end == -1 {
		return ""
	}
	return s[start : start+end]
}

func extractTextContent(s string) string {
	// Find first > and next <
	start := strings.IndexByte(s, '>')
	if start == -1 {
		return ""
	}
	start++
	end := strings.IndexByte(s[start:], '<')
	if end == -1 {
		return ""
	}
	return strings.TrimSpace(s[start : start+end])
}
