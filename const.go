package dsnet

// TODO allow env override (vars) as soon as someone needs it. Probably won't
// for ages, default interface name + locations are generally fine unless
// multiple VPNs are needed

const (
	INTERFACE_NAME = "dsnet"
	CONFIG_FILE = "/etc/dsnet-config.json"
	REPORT_FILE = "/var/lib/dsnet-report.json"
)
