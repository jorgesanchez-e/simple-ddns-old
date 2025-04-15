package ipify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/go-playground/validator/v10"
	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
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
	checkPeriodInMins int64
	urlIPV4           string
	urlIPV6           string
}

type ipifyConfig struct {
	CheckPeriodInMins int `yaml:"check-period-mins"`
	IPV4              struct {
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

	if err = yaml.Unmarshal(dataConfig, &config); err != nil {
		return nil, fmt.Errorf("unable to decode ipify config, err: %w", err)
	}

	return &publicIPs{
		checkPeriodInMins: int64(config.CheckPeriodInMins),
		urlIPV4:           config.IPV4.EndPoint,
		urlIPV6:           config.IPV6.EndPoint,
	}, nil
}

func (pip publicIPs) ipv4(ctx context.Context) *string {
	ip, err := pip.getIP(ctx, ipifyIPV4)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warning("unable to get ipv4")
		return nil
	}

	if err = validator.New().Var(ip, "ipv4"); err != nil {
		log.WithFields(log.Fields{"ipv4": ip, "error": err}).Warning("invalid ipv4")
		return nil
	}

	return &ip
}

func (pip publicIPs) ipv6(ctx context.Context) *string {
	ip, err := pip.getIP(ctx, ipifyIPV6)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warning("unable to get ipv6")
		return nil
	}

	if err = validator.New().Var(ip, "ipv6"); err != nil {
		log.WithFields(log.Fields{"ipv6": ip, "error": err}).Warning("invalid ipv6")
		return nil
	}

	return &ip
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

func (pip publicIPs) PublicIPs(ctx context.Context) ddns.IPs {
	return ddns.IPs{
		IPV4: pip.ipv4(ctx),
		IPV6: pip.ipv6(ctx),
	}
}

func (pip publicIPs) CheckPeriod() time.Duration {
	return time.Duration(pip.checkPeriodInMins * int64(time.Minute))
}
