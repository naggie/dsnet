package dsnet

import (
	"time"
)

const (
	// could be overridden in future via env
	CONFIG_FILE = "/etc/dsnetconfig.json"

	// these end up in the config file
	DEFAULT_INTERFACE_NAME = "dsnet"
	DEFAULT_REPORT_FILE    = "/var/lib/dsnetreport.json"
	DEFAULT_LISTEN_PORT    = 51820

	// keepalive always configured for everything. Set to a value likely to
	// stop most NATs from dropping the connection.
	KEEPALIVE = 21 * time.Second
	// allow missing a single keepalive + margin. Received data resets timeout, too.
	TIMEOUT = 50 * time.Second

	// when is a peer considered gone forever? (could remove)
	EXPIRY_DAYS = 28
)
