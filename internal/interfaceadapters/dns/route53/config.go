package route53

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v2"
)

type ConfigReader interface {
	Find(node string) (io.Reader, error)
}

type config struct {
	CredentialsFile *string `yaml:"credentials-file"`
	Records         []struct {
		FQDN            string
		Type            string
		ZoneID          string  `yaml:"zone-id"`
		CredentialsFile *string `yaml:"credentials-file"`
	}
}

func readConfig(reader ConfigReader) (config, error) {
	cnfReader, err := reader.Find(config_path)
	if err != nil {
		return config{}, fmt.Errorf("unable to get the config for %s node, err: %w", config_path, err)
	}

	dataConfig, err := io.ReadAll(cnfReader)
	if err != nil {
		return config{}, fmt.Errorf("unable to read route53 config, err: %w", err)
	}

	config := config{}
	if err = yaml.Unmarshal(dataConfig, &config); err != nil {
		return config, fmt.Errorf("unable to decode route53 config, err: %w", err)
	}

	return config, nil
}
