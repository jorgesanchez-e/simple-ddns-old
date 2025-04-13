package publicip

import (
	"encoding/json"
	"io"

	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/publicip/ipify"
)

const (
	ipify_service         string = "ipify"
	public_ip_config_path string = "ddns.public-ip-api"
)

type ConfigReader interface {
	Find(node string) (io.Reader, error)
}

func NewService(cnf ConfigReader) (ddns.PublicIPGetter, error) {
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
