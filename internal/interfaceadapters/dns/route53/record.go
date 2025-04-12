package route53

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/route53"
)

type records map[string]record

type record struct {
	fqdn       string
	recordType string
	zoneID     string
	client     *route53.Client
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
