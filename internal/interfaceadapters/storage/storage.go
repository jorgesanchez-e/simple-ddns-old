package storage

import (
	"encoding/json"
	"io"

	"github.com/jorgesanchez-e/simple-ddns/internal/domain/ddns"
	"github.com/jorgesanchez-e/simple-ddns/internal/interfaceadapters/storage/sqlite"
)

const (
	storage_config      string = "ddns.storage"
	sqlite_config       string = "ddns.storage.sqlite"
	sqlite_storage_name string = "sqlite"
)

type ConfigReader interface {
	Find(node string) (io.Reader, error)
}

func NewService(cnf ConfigReader) (ddns.Repository, error) {
	storageConfigReader, err := cnf.Find(storage_config)
	if err != nil {
		return nil, err
	}

	storageConfig, err := io.ReadAll(storageConfigReader)
	if err != nil {
		return nil, err
	}

	var config map[string]interface{}
	if err = json.Unmarshal(storageConfig, &config); err != nil {
		return nil, err
	}

	if _, exists := config[sqlite_storage_name]; exists {
		return sqlite.New(cnf)
	}

	return nil, nil
}
