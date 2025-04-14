package dns

import (
	"encoding/json"
	"io"
)

const (
	ddns_config_path          string = "ddns.dns-server"
	aws_dns_service           string = "ddns.dns-server.aws.route53"
	digital_ocean_dns_service string = "ddns.dns-server.digital-ocean.domain"
	aws_server_key            string = "aws"
	do_server_key             string = "digital-ocean.domain"
)

type ConfigReader interface {
	Find(node string) (io.Reader, error)
}

func readConfig(cnf ConfigReader) (map[string]interface{}, error) {
	cnfReader, err := cnf.Find(ddns_config_path)
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

	return config, nil
}
