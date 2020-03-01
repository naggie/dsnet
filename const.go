package dsnet

const (
	// could be overridden in future via env
	CONFIG_FILE = "/etc/dsnet-config.json"

	// these end up in the config file
	DEFAULT_INTERFACE_NAME = "dsnet"
	DEFAULT_REPORT_FILE    = "/var/lib/dsnet-report.json"
	DEFAULT_LISTEN_PORT    = 51820

	// keepalive always configured for everything
	KEEPALIVE_SECONDS = 21

	// when is a peer considered gone forever? (could remove)
	EXPIRY_DAYS = 28
)
