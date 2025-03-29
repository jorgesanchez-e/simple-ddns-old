package ddns

import "context"

type PublicIPs struct {
	IPV4 string
	IPV6 string
}

type Repository interface {
	Save(ctx context.Context, fqdn string, ip string) error
	Last(ctx context.Context, fqdn string) (ips PublicIPs, err error)
}
