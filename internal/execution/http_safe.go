package execution

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strconv"
	"time"
)

var blockedNetworks = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("2001:db8::/32"),
}

func publicIP(ip net.IP) bool {
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	if !address.IsGlobalUnicast() || address.IsPrivate() || address.IsLoopback() || address.IsLinkLocalUnicast() {
		return false
	}
	for _, network := range blockedNetworks {
		if network.Contains(address) {
			return false
		}
	}
	return true
}
func publicHTTPTransport() *http.Transport {
	return &http.Transport{
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(address)
			if err != nil {
				return nil, err
			}
			addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			if len(addresses) == 0 {
				return nil, errors.New("host has no addresses")
			}
			dialer := net.Dialer{Timeout: 10 * time.Second}
			for _, candidate := range addresses {
				if !publicIP(candidate.IP) {
					return nil, fmt.Errorf("requests to private or reserved addresses are blocked")
				}
			}
			var lastErr error
			for _, candidate := range addresses {
				connection, err := dialer.DialContext(ctx, network, net.JoinHostPort(candidate.IP.String(), port))
				if err == nil {
					return connection, nil
				}
				lastErr = err
			}
			return nil, lastErr
		},
	}
}
func validateHTTPURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
		return errors.New("HTTP node requires a complete http:// or https:// URL")
	}
	if parsed.User != nil {
		return errors.New("HTTP node URL must not contain credentials")
	}
	if parsed.Port() != "" {
		if _, err := strconv.Atoi(parsed.Port()); err != nil {
			return errors.New("HTTP node URL has an invalid port")
		}
	}
	return nil
}
