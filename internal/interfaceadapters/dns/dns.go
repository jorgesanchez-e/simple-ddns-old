package dns

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/dns/route53"
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

type recordManager interface {
	ddns.RecordUpdater
	DomainsManaged([]ddns.DNSRecord) []ddns.DNSRecord
}

type dnsUpdater struct {
	managers []recordManager
}

func NewService(ctx context.Context, cnf ConfigReader) (*dnsUpdater, error) {
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

	updater := dnsUpdater{managers: make([]recordManager, 0)}
	for key, _ := range config {
		switch key {
		case aws_server_key:
			dnsService, err := route53.New(ctx, cnf)
			if err != nil {
				return nil, err
			}
			updater.managers = append(updater.managers, dnsService)
		case do_server_key:
			continue
		}
	}

	return &updater, nil
}

func (du *dnsUpdater) Update(ctx context.Context, recs []ddns.DNSRecord) error {
	for _, manager := range du.managers {
		err := manager.Update(ctx, recs)
		if err != nil {
			return fmt.Errorf("unable to update records, err:%w", err)
		}
	}

	return nil
}
