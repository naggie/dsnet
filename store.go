package dsnet

import (
	"fmt"
)

const (
	JSON_FILE = "json_file"
	SQLITE    = "sqlite"
)

type Store interface {
	LoadDsnetConfig() (*DsnetConfig, error)
	StoreDsnetConfig(*DsnetConfig) error
}

func NewStore(storeType string, storeArgs map[string]string) (Store, error) {
	switch storeType {
	case JSON_FILE:
		newStore, err := NewFileStore(storeArgs)
		if err != nil {
			return nil, err
		}
		return newStore, nil
	case SQLITE:
		return nil, fmt.Errorf("sqlite store type unimplmented")
	}
	// If no storeType matched, then error out
	return nil, fmt.Errorf("%s is not a valid storage type", storeType)
}
