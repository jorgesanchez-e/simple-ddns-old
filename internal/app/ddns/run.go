package ddnsApp

import (
	"context"
	"os"

	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"

	log "github.com/sirupsen/logrus"
)

func (app *application) run(ctx context.Context, sig chan os.Signal, done chan bool) {
For:
	for {
		ctxIteration, cancel := context.WithTimeout(ctx, app.ips.CheckPeriod())

		select {
		case <-sig:
			cancel()
			break For
		case <-app.update(ctxIteration):
			cancel()
		}
	}

	done <- true
}

func (app *application) update(ctx context.Context) <-chan struct{} {
	recordsToUpdate := make([]ddns.DNSRecord, 0)

	ips := app.ips.PublicIPs(ctx)
	domains := app.updater.ManagedDomains()
	for index, domain := range domains {
		if domain.Type == ddns.IPV4DNSRec && ips.IPV4 != nil {
			domains[index].IP = *ips.IPV4
		}

		if domain.Type == ddns.IPV6DNSRec && ips.IPV6 != nil {
			domains[index].IP = *ips.IPV6
		}

		rec, err := app.storage.Last(ctx, domains[index])
		if err != nil {
			log.WithFields(log.Fields{"domain": domains[index].FQDN, "error": err}).Error("unable to check the database")
			continue
		}

		if rec != nil {
			recordsToUpdate = append(recordsToUpdate, *rec)
		}
	}

	if len(recordsToUpdate) > 0 {
		log.WithFields(log.Fields{"records": recordsToUpdate}).Info("the following records will be updated")
		if err := app.updater.Update(ctx, recordsToUpdate); err != nil {
			log.WithFields(log.Fields{"records": recordsToUpdate, "step": "dns"}).Errorf("unable to update records, err %s", err)
		}

		if err := app.storage.Update(ctx, recordsToUpdate); err != nil {
			log.WithFields(log.Fields{"records": recordsToUpdate, "step": "storage"}).Errorf("unable to update records, err %s", err)
		}
	} else {
		log.Info("all records are up to date")
	}

	return ctx.Done()
}
