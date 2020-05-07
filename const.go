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

	// keepalive always configured for clients. Set to a value likely to
	// stop most NATs from dropping the connection. Wireguard docs recommend 25
	// for most NATs
	KEEPALIVE = 25 * time.Second

	// if last handshake (different from keepalive, see https://www.wireguard.com/protocol/)
	TIMEOUT = 3 * time.Minute

	// when is a peer considered gone forever? (could remove)
	EXPIRY = 28 * time.Hour * 24
)

var (
	// populated with LDFLAGS, see do-release.sh
	VERSION = "unknown"
	GIT_COMMIT = "unknown"
	BUILD_DATE = "unknown"
)
