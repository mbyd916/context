// Package google provides a function to do Google searches using the Google Web
// Search API. See https://developers.google.com/web-search/docs/
//
// This package is an example to accompany https://blog.golang.org/context.
// It is not intended for use by others.
//
// Google has since disabled its search API,
// and so this package is no longer useful.
package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mbyd916/context/userip"
)

func main() {
	fmt.Println("vim-go")
}

// Results is an ordered list of search results.
type Results []Result

// A Result contains the title and URL of a search result.
type Result struct {
	Title, URL string
}

// Search sends query to Google search and returns the results.
func Search(ctx context.Context, query string) (Results, error) {
	// Prepare the Googlee Search API request.
	req, err := http.NewRequestWithContext(ctx, "GET", "https://ajax.googleapis.com/ajax/services/search/web?v=1.0", nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Set("q", query)

	// If ctx is carrying the user IP address, forward it to the server.
	// Google APIs use the user IP to distinguish server-initiated requests
	// from end-user requests.
	if userIP, ok := userip.FromContext(ctx); ok {
		q.Set("userip", userIP.String())
	}

	req.URL.RawQuery = q.Encode()

	// Issue the HTTP request and handle the response. The httpDo function
	// cancels the request if ctx.Done is closed.
	var results Results
	err = httpDo(ctx, req, func(rsp *http.Response, err error) error {
		if err != nil {
			return err
		}
		defer rsp.Body.Close()

		// Parse the JSON search result.
		var data struct {
			ResponseData struct {
				Results []struct {
					TitleNoFormatting string
					URL               string
				}
			}
		}
		if err := json.NewDecoder(rsp.Body).Decode(&data); err != nil {
			return err
		}
		for _, res := range data.ResponseData.Results {
			results = append(results, Result{Title: res.TitleNoFormatting, URL: res.URL})
		}
		return nil
	})

	return results, err
}

func httpDo(ctx context.Context, req *http.Request, f func(*http.Response, error) error) error {
	c := make(chan error, 1)
	go func() { c <- f(http.DefaultClient.Do(req)) }()

	select {
	case <-ctx.Done():
		<-c // Wait for f to return.
		return ctx.Err()
	case err := <-c:
		return err
	}
}
