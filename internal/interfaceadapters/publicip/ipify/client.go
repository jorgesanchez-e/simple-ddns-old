package ipify

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

const (
	ipifyIPV4 int = iota
	ipifyIPV6

	urlIPV4 string = "https://api.ipify.org"
	urlIPV6 string = "https://api6.ipify.org"
)

func getIP(ctx context.Context, ipType int) (_ string, err error) {
	url := ""

	switch ipType {
	case ipifyIPV4:
		url = urlIPV4
	case ipifyIPV6:
		url = urlIPV6
	default:
		return "", ErrInvalidIpType
	}

	req := &http.Request{}
	if req, err = http.NewRequest("GET", url, nil); err != nil {
		return "", fmt.Errorf("request build error: %w", err)
	}

	req = req.WithContext(ctx)
	res := &http.Response{}

	client := http.DefaultClient
	if res, err = client.Do(req); err != nil {
		return "", fmt.Errorf("network error: %w", err)
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http error: %w", fmt.Errorf("httpd code %d", res.StatusCode))
	}

	ip, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("response body error: %w", err)
	}

	return string(ip), nil
}
