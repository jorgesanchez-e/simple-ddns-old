package ddns

import "context"

const (
	IPV4DNSRec IPDNSRecType = "A"
	IPV6DNSRec IPDNSRecType = "AAAA"
)

type IPDNSRecType string

type DNSRecord struct {
	FQDN string
	Type IPDNSRecType
	IP   string
}

type RecordUpdater interface {
	Update(ctx context.Context, rec []DNSRecord) error
}
