package cli

import "fmt"

func Patch(patch map[string]interface{}) error {
	conf, err := LoadConfigFile()
	if err != nil {
		return fmt.Errorf("%w - failed to load config", err)
	}

	conf.Merge(patch)

	if err = conf.Save(); err != nil {
		return fmt.Errorf("%w - failure to save config", err)
	}

	return nil
}
