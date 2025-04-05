package ddns

import "context"

const (
	IPV4 IPType = "V4"
	IPV6 IPType = "V6"
)

type IPType string

type PublicIP struct {
	Value string
	Type  IPType
}

type PublicIPGetter interface {
	PublicIPs(ctx context.Context) ([]PublicIP, error)
}
