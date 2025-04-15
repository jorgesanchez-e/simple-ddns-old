package ddnsApp

import (
	"context"
	"fmt"
	"os"

	"github.com/jorgesanchez-e/simple-ddns/internal/config"
	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/dns"
	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/publicip"
	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/storage"

	log "github.com/sirupsen/logrus"
)

type application struct {
	updater ddns.RecordUpdater
	ips     ddns.PublicIPGetter
	storage ddns.Repository
}

func New(ctx context.Context) (_ *application, err error) {
	app := application{}

	defer func() {
		if err != nil {
			err = fmt.Errorf("unable to create ddns app, err:%w", err)
		}
	}()

	cnf, err := config.New()
	if err != nil {
		return nil, err
	}

	if app.storage, err = storage.NewService(cnf); err != nil {
		return nil, err
	}

	if app.ips, err = publicip.NewService(cnf); err != nil {
		return nil, err
	}

	if app.updater, err = dns.NewService(ctx, cnf); err != nil {
		return nil, err
	}

	return &app, nil
}

func (app *application) Run(ctx context.Context) chan bool {
	done := make(chan bool, 1)
	sig := make(chan os.Signal, 1)

	log.Info("running")
	go app.run(ctx, sig, done)

	return done
}
