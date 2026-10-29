package cli

import "fmt"

// SetPaused pauses or resumes a peer by hostname, then syncs. A paused peer
// stays in the config but is removed from the wg interface.
func SetPaused(hostname string, paused bool) error {
	conf, err := LoadConfigFile()
	if err != nil {
		return fmt.Errorf("%w - failed to load config", err)
	}

	found := false
	for i := range conf.Peers {
		if conf.Peers[i].Hostname == hostname {
			conf.Peers[i].Paused = paused
			found = true
		}
	}

	if !found {
		return fmt.Errorf("failed to find peer with hostname %s", hostname)
	}

	if err = conf.Save(); err != nil {
		return fmt.Errorf("%w - failure to save config", err)
	}
	server := GetServer(conf)

	if err = server.ConfigureDevice(); err != nil {
		return fmt.Errorf("%w - failed to sync server config to wg interface: %s", err, server.InterfaceName)
	}
	return nil
}
