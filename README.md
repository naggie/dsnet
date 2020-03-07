dsnet is a simple tool to manage a centralised wireguard VPN. Think wg-quick
but quicker.

    Usage: dsnet <cmd>

    Available commands:

    	init   : Create /etc/dsnetconfig.json containing default configuration + new keys without loading. Edit to taste.
    	add    : Add a new peer + sync
    	up     : Create the interface, run pre/post up, sync
    	report : Generate a JSON status report to the location configured in /etc/dsnetconfig.json.
    	remove : Remove a peer by hostname provided as argument + sync
    	down   : Destroy the interface, run pre/post down
    	sync   : Update wireguard configuration from /etc/dsnetconfig.json after validating


Quick start -- install wireguard and dsnet, then:

    sudo dsnet init
    sudo dsnet up
    # edit /etc/dsnetconfig.json to taste
	dsnet add banana > dsnet-banana.conf
	dsnet add apple > dsnet-apple.conf

Copy the configuration file to your devices and connect!

Dsnet assumes a DNS server is running on the server at the moment.

To send configurations, ffsend (with separately transferred password) or a local QR code generator may be used.

TODO after first release:

  * Hooks for adding routes/ IPtables forwarding rules
  * Route entire internet option
  * Support for additional subnets in peer config (with routes)
  * Peer endpoint support
