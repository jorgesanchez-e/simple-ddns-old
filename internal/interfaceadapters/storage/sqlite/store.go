package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
	_ "github.com/mattn/go-sqlite3"
)

const (
	ipv4Regexp    string = `\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}`
	sqlite_config string = "ddns.storage.sqlite"
)

var (
	once sync.Once
	str  = &store{}
)

type ConfigReader interface {
	Find(node string) (io.Reader, error)
}

type store struct {
	driver *sql.DB
}

type sqliteConfig struct {
	DB string `json:"db"`
}

func New(cnf ConfigReader) (*store, error) {
	var err error

	cnfReader, err := cnf.Find(sqlite_config)
	if err != nil {
		return nil, fmt.Errorf("sqlite config not found, err: %w", err)
	}

	dataConfig, err := io.ReadAll(cnfReader)
	if err != nil {
		return nil, fmt.Errorf("unable to read sqlite config, err: %w", err)
	}

	config := sqliteConfig{}
	if err = json.Unmarshal(dataConfig, &config); err != nil {
		return nil, fmt.Errorf("unable to decode sqlite config, err: %w", err)
	}

	db := &sql.DB{}

	once.Do(func() {
		db, err = sql.Open("sqlite3", config.DB)
		if err == nil {
			str.driver = db
		}
	})

	if err != nil {
		return nil, fmt.Errorf("unable to open sqlite database %s, err: %w", config.DB, err)
	}

	_, err = db.Exec(createTable)
	if err != nil {
		return nil, fmt.Errorf("unable to execute sql statement %s, err: %w", createTable, err)
	}

	return str, err
}

func (d *store) Update(ctx context.Context, records []ddns.DNSRecord) error {
	for _, record := range records {
		registerType, err := getDNSRegType(record.FQDN, record.IP)
		if err != nil {
			return fmt.Errorf("unable to get register type, fqdn:%s, ip:%s, err: %w", record.FQDN, record.IP, err)
		}

		if err = d.saveRecord(ctx, record.FQDN, record.IP, registerType); err != nil {
			return fmt.Errorf("unable to save register, fqdn:%s, ip:%s, err: %w", record.FQDN, record.IP, err)
		}
	}

	return nil
}

func (d *store) Last(ctx context.Context, fqdn string) (*ddns.DNSRecord, error) {
	registerTypes := []string{string(ddns.IPV4DNSRec), string(ddns.IPV6DNSRec)}
	record := &ddns.DNSRecord{}

	for index := 0; index < len(registerTypes); index++ {
		row := d.driver.QueryRowContext(
			ctx,
			lastRegister,
			fqdn,
			registerTypes[index],
		)

		ip := ""
		recType := ""

		err := row.Scan(&ip, &recType)
		if err != nil && err != sql.ErrNoRows {
			return nil, fmt.Errorf("unable to read register, ip:%s, err: %w", ip, err)
		}

		record.IP = ip
		record.FQDN = fqdn
		record.Type = ddns.IPDNSRecType(recType)
	}

	return record, nil
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

func getDNSRegType(fqdn, ip string) (string, error) {
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
		registerType = string(ddns.IPV4DNSRec)
	} else {
		if err = validator.Var(ip, "ipv6"); err != nil {
			return "", fmt.Errorf("invalid ipv6[%s], error: %w", ip, err)
		}
		registerType = string(ddns.IPV6DNSRec)
	}

	return registerType, nil
}
