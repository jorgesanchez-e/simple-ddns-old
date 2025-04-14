package route53

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
)

type batch struct {
	client       *route53.Client
	route53Batch []types.ChangeBatch
	zoneID       string
}

func (r53 *updater) batches(changes []ddns.DNSRecord) []batch {
	batches := make([]batch, 0, r53.recs.numberOfBatches())

	for _, rec2update := range changes {
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
