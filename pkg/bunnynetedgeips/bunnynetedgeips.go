package bunnynetedgeips

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PatrickCronin/routesum/pkg/routesum"
)

const (
	ipv4url = "https://api.bunny.net/system/edgeserverlist"
	ipv6url = ipv4url + "/ipv6"
)

func fetch(ctx context.Context, url string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("http request %s failed with code: %d", url, resp.StatusCode)
	}
	var ips []string
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&ips)
	if err != nil {
		return nil, fmt.Errorf("failed to decode json for %s: %w", url, err)
	}
	return ips, nil
}

func getIPs(ctx context.Context) ([]string, error) {
	ipsv4, err := fetch(ctx, ipv4url)
	if err != nil {
		return nil, err
	}
	ipsv6, err := fetch(ctx, ipv6url)
	if err != nil {
		return nil, err
	}
	return append(ipsv4, ipsv6...), nil
}

func BunnynetEdgeIPs(ctx context.Context) ([]string, error) {
	ips, err := getIPs(ctx)
	if err != nil {
		return nil, err
	}
	rs := routesum.NewRouteSum()
	for _, ip := range ips {
		err = rs.InsertFromString(ip)
		if err != nil {
			return nil, err
		}
	}
	return rs.SummaryStrings(), nil
}
