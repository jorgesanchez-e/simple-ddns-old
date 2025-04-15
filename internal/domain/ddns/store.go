package ddns

import "context"

type Repository interface {
	Update(ctx context.Context, record []DNSRecord) error
	Last(ctx context.Context, rec DNSRecord) (*DNSRecord, error)
}
