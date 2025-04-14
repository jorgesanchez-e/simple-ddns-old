package dns

import (
	"context"
	"fmt"

	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/dns/route53"
)

type dnsUpdater struct {
	managers []ddns.RecordUpdater
}

func NewService(ctx context.Context, cnf ConfigReader) (*dnsUpdater, error) {
	config, err := readConfig(cnf)
	if err != nil {
		return nil, err
	}

	updater := dnsUpdater{
		managers: make([]ddns.RecordUpdater, 0),
	}
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
