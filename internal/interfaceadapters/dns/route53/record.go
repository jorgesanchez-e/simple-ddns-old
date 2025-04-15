package route53

import (
	"context"
	"fmt"

	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
)

type records map[string]record

type record struct {
	dnsRec ddns.DNSRecord
	zoneID string
	client *route53.Client
}

func (r53 *updater) getManagedRecords(ctx context.Context, cnf config) error {
	managedRecords := make(records, 0)
	for _, configRecord := range cnf.Records {
		key := fmt.Sprintf("%s:%s", configRecord.FQDN, configRecord.Type)
		managedRecords[key] = record{
			dnsRec: ddns.DNSRecord{
				FQDN: configRecord.FQDN,
				Type: ddns.IPDNSRecType(configRecord.Type),
			},
			zoneID: configRecord.ZoneID,
		}

		credentialFiles := make([]string, 0)
		if configRecord.CredentialsFile != nil {
			credentialFiles = append(credentialFiles, *configRecord.CredentialsFile)
			awsCnf, err := awsConfig.LoadDefaultConfig(
				ctx,
				awsConfig.WithSharedConfigFiles(credentialFiles),
			)

			if err != nil {
				return fmt.Errorf("unable to get aws config for domain %s, error: %w", configRecord.FQDN, err)
			}

			r53Driver := route53.NewFromConfig(awsCnf)
			if r53Driver == nil {
				return fmt.Errorf("unable to create route53 connection, error: %w", err)
			}

			managedRecords.addClient(configRecord.FQDN, configRecord.Type, r53Driver)
		}
	}

	r53.recs = managedRecords

	return nil
}

func (mr records) find(fqdn, recordType string) *record {
	if rec, exists := mr[fmt.Sprintf("%s:%s", fqdn, recordType)]; exists {
		return &rec
	}

	return nil
}

func (mr records) addClient(fqdn, recordType string, client *route53.Client) {
	key := fmt.Sprintf("%s:%s", fqdn, recordType)
	if rec, exists := mr[key]; exists {
		rec.client = client
		mr[key] = rec
	}
}

func (mr records) numberOfBatches() int {
	recordsWithClient := 0
	for _, rec := range mr {
		if rec.client != nil {
			recordsWithClient++
		}
	}

	return recordsWithClient + 1
}

func (mr records) getDomains() []string {
	domains := make([]string, 0, len(mr))
	for key, _ := range mr {
		domains = append(domains, key)
	}

	return domains
}
