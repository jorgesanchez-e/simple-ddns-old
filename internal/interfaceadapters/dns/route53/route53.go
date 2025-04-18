package route53

import (
	"context"
	"fmt"

	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

const (
	config_path = "ddns.dns-server.aws.route53"
)

type updater struct {
	globalClient *route53.Client
	recs         records
}

func New(ctx context.Context, cnf ConfigReader) (*updater, error) {
	config, err := readConfig(cnf)
	if err != nil {
		return nil, err
	}

	var globalClient *route53.Client
	if config.CredentialsFile != nil {
		awsCnf, err := awsConfig.LoadDefaultConfig(
			ctx,
			awsConfig.WithSharedConfigFiles(
				[]string{*config.CredentialsFile},
			),
		)

		if err != nil {
			return nil, fmt.Errorf("unable to get aws config, err:%w", err)
		}

		globalClient = route53.NewFromConfig(awsCnf)
	}

	r53Updater := updater{
		globalClient: globalClient,
	}

	return &r53Updater, r53Updater.getManagedRecords(ctx, config)
}

func (r53 *updater) Update(ctx context.Context, recs []ddns.DNSRecord) error {
	domains := r53.recs.getDomains()
	if len(domains) == 0 {
		return nil
	}

	batches := r53.batches(recs)
	for _, batch := range batches {
		if batch.client == nil {
			_, err := r53.globalClient.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
				ChangeBatch:  &batch.route53Batch[0],
				HostedZoneId: aws.String(batch.zoneID),
			})

			if err != nil {
				return fmt.Errorf("unable to update the record, err:%w", err)
			}
		}
	}

	return nil
}

func (r53 *updater) ManagedDomains() []ddns.DNSRecord {
	records := make([]ddns.DNSRecord, 0)
	for _, value := range r53.recs {
		records = append(records, value.dnsRec)
	}
	return records
}
