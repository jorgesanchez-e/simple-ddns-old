package ddns

import "context"

type Repository interface {
	Save(ctx context.Context, domain Domain) error
	Last(ctx context.Context, fqdn FQDN) (Domain, error)
}
