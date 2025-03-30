package publicip

import (
	"context"
	"encoding/json"
	"io"

	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/publicip/ipify"
)

const (
	ipify_service         string = "ipify"
	public_ip_config_path string = "app.public-ip-api"
)

type ConfigReader interface {
	Find(node string) (io.Reader, error)
}

type PublicIPs interface {
	IPV4(ctx context.Context) (string, error)
	IPV6(ctx context.Context) (string, error)
}

func NewService(cnf ConfigReader) (PublicIPs, error) {
	cnfReader, err := cnf.Find(public_ip_config_path)
	if err != nil {
		return nil, err
	}

	configData, err := io.ReadAll(cnfReader)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}

	if err = json.Unmarshal(configData, &config); err != nil {
		return nil, err
	}

	if _, exists := config[ipify_service]; exists {
		return ipify.New(cnf)
	}

	return nil, nil
}
