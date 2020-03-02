dsnet is a simple tool to manage a wireguard VPN. Think wg-quick but quicker.

    Usage: dsnet <cmd>

    Available commands:

    	init   : Create /etc/dsnet-config.json containing default configuration + new keys without loading. Edit to taste.
    	add    : Generate configuration for a new peer, adding to /etc/dsnet-config.json. Send with passworded ffsend.
    	sync   : Synchronise wireguard configuration with /etc/dsnet-config.json, creating and activating interface if necessary.
    	report : Generate a JSON status report to the location configured in /etc/dsnet-config.json.

    To remove an interface or bring it down, use standard tools such as iproute2.
    To modify or remove peers, edit /etc/dsnet-config.json and then run sync.



To send configurations, ffsend (with separately transferred password) or a local QR code generator may be used.

TODO after first release:

  * Hooks for adding routes/ IPtables forwarding rules
  * Forward option
  * Support for additional subnets in peer config
  * Peer endpoint support
