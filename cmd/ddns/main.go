package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jorgesanchez-e/simple-ddns/internal/config"
	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/publicip"
	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/storage"
)

func main() {
	ctx := context.Background()

	cnf, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	repository, err := storage.NewService(cnf)
	if err != nil {
		log.Fatal(err)
	}

	pip, err := publicip.NewService(cnf)
	if err != nil {
		log.Fatal(err)
	}

	ipv4 := ""
	if ipv4, err = pip.IPV4(ctx); err != nil {
		log.Printf("unable to get ipv4, err:%s", err)
	}

	ipv6 := ""
	if ipv6, err = pip.IPV6(ctx); err != nil {
		log.Printf("unable to get ipv6, err:%s", err)
	}

	records := make([]ddns.DNSRecord, 0, 2)
	if ipv4 != "" {
		records = append(records, ddns.DNSRecord{
			FQDN: "vpn.jorgesanchez-e.dev",
			IP:   ipv4,
			Type: ddns.IPV4DNSRec,
		})
	}

	if ipv6 != "" {
		records = append(records, ddns.DNSRecord{
			FQDN: "vpn6.jorgesanchez-e.dev",
			IP:   ipv6,
			Type: ddns.IPV6DNSRec,
		})
	}

	if err = repository.Update(ctx, records); err != nil {
		log.Fatal(err)
	}

	rec, err := repository.Last(ctx, "home6.jorgesanchez-e.dev")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("fqdn:%s ip:%s type:%s\n", rec.FQDN, rec.IP, rec.Type)
}
