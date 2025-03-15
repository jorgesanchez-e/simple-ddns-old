package ddns

import "context"

type PublicIP interface {
	IPV4(ctx context.Context) (string, error)
	IPV6(ctx context.Context) (string, error)
}
