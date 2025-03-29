package ddns

import "context"

var (
	IPV4Type RegType = "A"
	IPV6Type RegType = "AAAA"
)

type RegType string

type Domain struct {
	FQDN string
	IP   string
}

type Updater interface {
	Update(ctx context.Context, domain Domain) error
}
