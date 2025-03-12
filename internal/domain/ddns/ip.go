package ddns

import "context"

type IP interface {
	String() string
	IsEqual(IP) bool
}

type PublicIP interface {
	IPV4(ctx context.Context) (IP, error)
	IPV6(ctx context.Context) (IP, error)
}
