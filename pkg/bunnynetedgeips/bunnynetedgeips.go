package bunnynetedgeips

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/PatrickCronin/routesum/pkg/routesum"
)

const (
	ipv4url         = "https://api.bunny.net/system/edgeserverlist"
	ipv6url         = ipv4url + "/ipv6"
	defaultTimeout  = 10 * time.Second
	maxResponseSize = 10 << 20 // 10 MB
)

type ipResponse struct {
	ips []string
	err error
}

func fetch(ctx context.Context, url string) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	client := &http.Client{
		Timeout: defaultTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}
	bodyReader := io.LimitReader(resp.Body, maxResponseSize)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}
	var ips []string
	if err := json.Unmarshal(body, &ips); err != nil {
		return nil, fmt.Errorf("decoding JSON from %s: %w", url, err)
	}
	return ips, nil
}

func getIPs(ctx context.Context) ([]string, error) {
	var wg sync.WaitGroup
	responses := make(chan ipResponse, 2)
	for _, url := range []string{ipv4url, ipv6url} {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			ips, err := fetch(ctx, url)
			responses <- ipResponse{ips: ips, err: err}
		}(url)
	}
	go func() {
		wg.Wait()
		close(responses)
	}()
	var allIPs []string
	for resp := range responses {
		if resp.err != nil {
			return nil, resp.err
		}
		allIPs = append(allIPs, resp.ips...)
	}
	return allIPs, nil
}

func BunnynetEdgeIPs(ctx context.Context) ([]string, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ips, err := getIPs(ctx)
	if err != nil {
		return nil, fmt.Errorf("fetching IPs: %w", err)
	}
	rs := routesum.NewRouteSum()
	for _, ip := range ips {
		if err := rs.InsertFromString(ip); err != nil {
			return nil, fmt.Errorf("inserting IP %s: %w", ip, err)
		}
	}
	return rs.SummaryStrings(), nil
}
