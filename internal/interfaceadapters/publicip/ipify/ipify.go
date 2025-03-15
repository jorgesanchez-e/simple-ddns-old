package ipify

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var ErrInvalidIpType = errors.New("invalid ip type argument")

type publicIPs struct {
}

func New() *publicIPs {
	return &publicIPs{}
}

func (pip publicIPs) IPV4(ctx context.Context) (string, error) {
	ip, err := getIP(ctx, ipifyIPV4)
	if err != nil {
		return "", err
	}

	if err = validator.New().Var(ip, "ipv4"); err != nil {
		return "", fmt.Errorf("invalid ipv4 [%s], format error: %w", ip, err)
	}

	return ip, nil
}

func (pip publicIPs) IPV6(ctx context.Context) (string, error) {
	ip, err := getIP(ctx, ipifyIPV6)
	if err != nil {
		return "", err
	}

	if err = validator.New().Var(ip, "ipv6"); err != nil {
		return "", fmt.Errorf("invalid ipv6 [%s], format error: %w", ip, err)
	}

	return ip, nil
}
