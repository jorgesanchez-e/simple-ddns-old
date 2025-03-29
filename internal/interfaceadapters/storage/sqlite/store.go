package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ipv4Regexp string = `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`
)

var (
	once sync.Once
	str  = &store{}
)

type store struct {
	driver *sql.DB
}

func New(ctx context.Context) (*store, error) {
	var err error
	db := &sql.DB{}

	once.Do(func() {
		db, err = sql.Open("sqlite3", "./app.db")
		if err == nil {
			str.driver = db
		}
	})

	if err != nil {
		return nil, err
	}

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	return str, err
}

func (d *store) Save(ctx context.Context, fqdn string, ip string) error {
	registerType, err := getRegisterType(fqdn, ip)
	if err != nil {
		return err
	}

	return d.saveRecord(ctx, fqdn, ip, registerType)
}

func (d *store) Last(ctx context.Context, fqdn string) (*ddns.PublicIPs, error) {
	registerTypes := []string{string(ddns.IPV4Type), string(ddns.IPV6Type)}
	publicIPs := ddns.PublicIPs{}

	for index := 0; index < len(registerTypes); index++ {
		row := d.driver.QueryRowContext(
			ctx,
			lastRegister,
			fqdn,
			registerTypes[index],
		)

		ip := ""

		err := row.Scan(&ip)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}

		if registerTypes[index] == string(ddns.IPV4Type) {
			publicIPs.IPV4 = ip
		} else {
			publicIPs.IPV6 = ip
		}
	}

	return &publicIPs, nil
}

func (d *store) saveRecord(ctx context.Context, fqdn string, ip string, rType string) error {
	tx, err := d.driver.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("transaction error: %w", err)
	}

	if _, err = tx.Exec(updateRegister, fqdn, rType); err != nil {
		tx.Rollback()
		return fmt.Errorf("update error: %w", err)
	}

	if _, err = tx.Exec(insertRegister, fqdn, time.Now(), rType, ip, true); err != nil {
		tx.Rollback()
		return fmt.Errorf("insert error: %w", err)
	}

	tx.Commit()

	return nil
}

func getRegisterType(fqdn, ip string) (string, error) {
	var err error
	validator := validator.New()
	registerType := ""

	if err = validator.Var(fqdn, "fqdn"); err != nil {
		return "", fmt.Errorf("invalid fqdn [%s], format error: %w", fqdn, err)
	}

	matched, err := regexp.MatchString(ipv4Regexp, ip)
	if err != nil {
		return "", fmt.Errorf("invalid ipv4 regexp [%s], error: %w", fqdn, err)
	}

	if matched {
		if err = validator.Var(ip, "ipv4"); err != nil {
			return "", fmt.Errorf("invalid ipv4 [%s], error: %w", ip, err)
		}
		registerType = string(ddns.IPV4Type)
	} else {
		if err = validator.Var(ip, "ipv6"); err != nil {
			return "", fmt.Errorf("invalid ipv6[%s], error: %w", ip, err)
		}
		registerType = string(ddns.IPV6Type)
	}

	return registerType, nil
}
