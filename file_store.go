package dsnet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-playground/validator"
)

type FileStore struct {
	path string
}

func NewFileStore(args map[string]string) (Store, error) {
	path, ok := args["path"]
	if !ok {
		return &FileStore{}, fmt.Errorf("missing required argument path")
	}
	return &FileStore{
		path: path,
	}, nil
}

func (f FileStore) LoadDsnetConfig() (*DsnetConfig, error) {
	raw, err := ioutil.ReadFile(f.path)

	if os.IsNotExist(err) {
		return nil, fmt.Errorf("%s does not exist. `dsnet init` may be required.", f.path)
	} else if os.IsPermission(err) {
		return nil, fmt.Errorf("%s cannot be accessed. Sudo may be required.", f.path)
	} else if err != nil {
		return nil, err
	}

	conf := DsnetConfig{}
	err = json.Unmarshal(raw, &conf)
	if err != nil {
		return nil, err
	}

	err = validator.New().Struct(conf)
	if err != nil {
		return nil, err
	}

	if conf.ExternalHostname == "" && len(conf.ExternalIP) == 0 && len(conf.ExternalIP6) == 0 {
		return nil, fmt.Errorf("Config does not contain ExternalIP, ExternalIP6 or ExternalHostname")
	}

	return &conf, nil
}

func (f FileStore) StoreDsnetConfig(conf *DsnetConfig) error {
	_json, _ := json.MarshalIndent(conf, "", "    ")
	err := ioutil.WriteFile(f.path, _json, 0600)
	if err != nil {
		return err
	}
	return nil
}
