package ipify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
)

const (
	config_path     = "ddns.public-ip-api.ipify"
	ipifyIPV4   int = iota
	ipifyIPV6
)

var ErrInvalidIpType = errors.New("invalid ip type argument")

type ConfigReader interface {
	Find(node string) (io.Reader, error)
}

type publicIPs struct {
	urlIPV4 string
	urlIPV6 string
}

type ipifyConfig struct {
	IPV4 struct {
		EndPoint string
	}
	IPV6 struct {
		EndPoint string
	}
}

func New(cnf ConfigReader) (*publicIPs, error) {
	config := ipifyConfig{}

	cnfReader, err := cnf.Find(config_path)
	if err != nil {
		return nil, fmt.Errorf("ipify config not found, err: %w", err)
	}

	dataConfig, err := io.ReadAll(cnfReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read ipify config, err: %w", err)
	}

	if err = json.Unmarshal(dataConfig, &config); err != nil {
		return nil, fmt.Errorf("unable to decode ipify config, err: %w", err)
	}

	return &publicIPs{
		urlIPV4: config.IPV4.EndPoint,
		urlIPV6: config.IPV6.EndPoint,
	}, nil
}

func (pip publicIPs) IPV4(ctx context.Context) (string, error) {
	ip, err := pip.getIP(ctx, ipifyIPV4)
	if err != nil {
		return "", fmt.Errorf("get ipv4 ip error, err: %w", err)
	}

	if err = validator.New().Var(ip, "ipv4"); err != nil {
		return "", fmt.Errorf("invalid ipv4 [%s], format error: %w", ip, err)
	}

	return ip, nil
}

func (pip publicIPs) IPV6(ctx context.Context) (string, error) {
	ip, err := pip.getIP(ctx, ipifyIPV6)
	if err != nil {
		return "", fmt.Errorf("get ipv6 ip error, err: %w", err)
	}

	if err = validator.New().Var(ip, "ipv6"); err != nil {
		return "", fmt.Errorf("invalid ipv6 [%s], format error: %w", ip, err)
	}

	return ip, nil
}

func (pip publicIPs) getIP(ctx context.Context, ipType int) (_ string, err error) {
	url := ""

	switch ipType {
	case ipifyIPV4:
		url = pip.urlIPV4
	case ipifyIPV6:
		url = pip.urlIPV6
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
