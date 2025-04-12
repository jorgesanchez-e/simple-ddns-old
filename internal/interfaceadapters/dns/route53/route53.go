package route53

import (
	"context"
	"fmt"
	"io"

	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
	"gopkg.in/yaml.v2"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

const (
	config_path = "ddns.dns-server.aws.route53"
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

type updater struct {
	globalConnection *route53.Client
	recs             records
}

type batch struct {
	client       *route53.Client
	route53Batch []types.ChangeBatch
	zoneID       string
}

func New(ctx context.Context, cnf ConfigReader) (*updater, error) {
	cnfReader, err := cnf.Find(config_path)
	if err != nil {
		return nil, fmt.Errorf("unable to get the config for %s node, err: %w", config_path, err)
	}

	dataConfig, err := io.ReadAll(cnfReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read route53 config, err: %w", err)
	}

	config := config{}
	if err = yaml.Unmarshal(dataConfig, &config); err != nil {
		return nil, fmt.Errorf("unable to decode route53 config, err: %w", err)
	}

	r53Updater := updater{}

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
	r53Updater.globalConnection = globalClient

	return &r53Updater, r53Updater.getManagedRecords(ctx, config)
}

func (r53 *updater) getManagedRecords(ctx context.Context, cnf config) error {
	managedRecords := make(records, 0)
	for _, configRecord := range cnf.Records {
		key := fmt.Sprintf("%s:%s", configRecord.FQDN, configRecord.Type)
		managedRecords[key] = record{
			fqdn:       configRecord.FQDN,
			recordType: configRecord.Type,
			zoneID:     configRecord.ZoneID,
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

func (r53 *updater) Update(ctx context.Context, recs []ddns.DNSRecord) error {
	managedRecords := r53.DomainsManaged(recs)
	if len(managedRecords) == 0 {
		return nil
	}

	for index, batch := range r53.batches(managedRecords) {
		if batch.client == nil {
			output, err := r53.globalConnection.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
				ChangeBatch:  &batch.route53Batch[index],
				HostedZoneId: aws.String(batch.zoneID),
			})

			if err != nil {
				return fmt.Errorf("unable to update the record, err:%w", err)
			}
			fmt.Printf("%v", output)
		}
	}

	return nil
}

func (r53 *updater) DomainsManaged(toUpdate []ddns.DNSRecord) []ddns.DNSRecord {
	managedRecords := []ddns.DNSRecord{}
	for _, rec := range toUpdate {
		for _, r53Rec := range r53.recs {
			if r53Rec.fqdn == rec.FQDN && r53Rec.recordType == string(rec.Type) {
				managedRecords = append(managedRecords, rec)
			}
		}
	}

	return managedRecords
}

func (r53 *updater) batches(toUpdate []ddns.DNSRecord) []batch {
	batches := make([]batch, 0, r53.recs.numberOfBatches())

	for _, rec2update := range toUpdate {
		rec := r53.recs.find(rec2update.FQDN, string(rec2update.Type))
		if rec == nil {
			continue
		}

		batches = append(batches, batch{
			client: rec.client,
			route53Batch: []types.ChangeBatch{
				{
					Changes: []types.Change{
						{
							Action: types.ChangeActionUpsert,
							ResourceRecordSet: &types.ResourceRecordSet{
								Name: aws.String(rec.fqdn),
								Type: route53RecordType(rec.recordType),
								TTL:  aws.Int64(300),
								ResourceRecords: []types.ResourceRecord{
									{
										Value: aws.String(rec2update.IP),
									},
								},
							},
						},
					},
				},
			},
			zoneID: rec.zoneID,
		})
	}

	return batches
}

func route53RecordType(rType string) types.RRType {
	switch ddns.IPDNSRecType(rType) {
	case ddns.IPV4DNSRec:
		return types.RRTypeA
	case ddns.IPV6DNSRec:
		return types.RRTypeAaaa
	default:
		return types.RRTypeA
	}
}
