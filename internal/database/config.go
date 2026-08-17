package database

import (
	"os"
	"path/filepath"
)

const (
	DEFAULT_DB_FOLDER = ".floppy"
)

type DatabaseConfig struct {
	Path string
}

func (dbc *DatabaseConfig) setDefaults() error {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dbc.Path = filepath.Join(homeDir, DEFAULT_DB_FOLDER)

	return nil
}

func newDBC(opts ...Option) (*DatabaseConfig, error) {
	dbc := new(DatabaseConfig)

	err := dbc.setDefaults()

	if err != nil {
		return nil, err
	}

	for _, o := range opts {
		o(dbc)
	}

	return dbc, nil

}
