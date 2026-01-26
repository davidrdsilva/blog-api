package services

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/davidrdsilva/blog-api/internal/application/dtos"
	"golang.org/x/net/html"
)

// URLService handles URL metadata fetching
type URLService struct{}

// NewURLService creates a new URL service
func NewURLService() *URLService {
	return &URLService{}
}

// FetchURLMetadata fetches metadata from a URL for Editor.js Link Tool
func (s *URLService) FetchURLMetadata(targetURL string) (*dtos.EditorJsURLResponse, error) {
	// Validate URL
	parsedURL, err := url.Parse(targetURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return &dtos.EditorJsURLResponse{
			Success: 0,
			Error: &dtos.EditorJsErrorDetail{
				Code:    "INVALID_URL",
				Message: "URL parameter is missing or malformed",
			},
		}, nil
	}

	// Create HTTP client with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	if err != nil {
		return &dtos.EditorJsURLResponse{
			Success: 0,
			Error: &dtos.EditorJsErrorDetail{
				Code:    "URL_NOT_ACCESSIBLE",
				Message: "Unable to fetch metadata from the provided URL",
			},
		}, nil
	}

	// Set user agent to avoid some anti-bot protections
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; BlogAPI/1.0; +http://example.com/bot)")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		errCode := "URL_NOT_ACCESSIBLE"
		if ctx.Err() == context.DeadlineExceeded {
			errCode = "REQUEST_TIMEOUT"
		}
		return &dtos.EditorJsURLResponse{
			Success: 0,
			Error: &dtos.EditorJsErrorDetail{
				Code:    errCode,
				Message: fmt.Sprintf("Unable to fetch the URL: %v", err),
			},
		}, nil
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return &dtos.EditorJsURLResponse{
			Success: 0,
			Error: &dtos.EditorJsErrorDetail{
				Code:    "URL_NOT_ACCESSIBLE",
				Message: fmt.Sprintf("URL returned status code: %d", resp.StatusCode),
			},
		}, nil
	}

	// Parse HTML to extract metadata
	metadata, err := s.extractMetadata(resp.Body)
	if err != nil {
		return &dtos.EditorJsURLResponse{
			Success: 0,
			Error: &dtos.EditorJsErrorDetail{
				Code:    "PARSE_ERROR",
				Message: "Failed to parse URL metadata",
			},
		}, nil
	}

	return &dtos.EditorJsURLResponse{
		Success: 1,
		Link:    targetURL,
		Meta:    metadata,
	}, nil
}

// extractMetadata parses HTML and extracts Open Graph and meta tags
func (s *URLService) extractMetadata(body io.Reader) (*dtos.URLMetadata, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	metadata := &dtos.URLMetadata{}
	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				// Extract title text
				if n.FirstChild != nil && metadata.Title == "" {
					metadata.Title = n.FirstChild.Data
				}
			case "meta":
				// Extract meta tags
				var property, name, content string
				for _, attr := range n.Attr {
					switch attr.Key {
					case "property":
						property = attr.Val
					case "name":
						name = attr.Val
					case "content":
						content = attr.Val
					}
				}

				// Open Graph tags take priority
				switch property {
				case "og:title":
					if content != "" {
						metadata.Title = content
					}
				case "og:description":
					if content != "" {
						metadata.Description = content
					}
				case "og:image":
					if content != "" {
						metadata.Image = &dtos.URLImageInfo{URL: content}
					}
				}

				// Fallback to standard meta tags
				if name == "description" && metadata.Description == "" && content != "" {
					metadata.Description = content
				}
				if (name == "twitter:image" || property == "twitter:image") && metadata.Image == nil && content != "" {
					metadata.Image = &dtos.URLImageInfo{URL: content}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	// Trim whitespace from extracted values
	metadata.Title = strings.TrimSpace(metadata.Title)
	metadata.Description = strings.TrimSpace(metadata.Description)

	return metadata, nil
}
