package ddns

import "context"

type FQDN interface {
	Name() string
	RegisterType() string
}

type Domain interface {
	FQDN
	IP() string
}

type Updater interface {
	Update(ctx context.Context, domain Domain) error
}
