package ddns

import (
	"context"
	"time"
)

type IPs struct {
	IPV4 *string
	IPV6 *string
}

type PublicIPGetter interface {
	PublicIPs(ctx context.Context) IPs
	CheckPeriod() time.Duration
}
