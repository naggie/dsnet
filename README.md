Purpose: to allow an overview of Darksky peers and provide a mechanism to allow
easy joining



OUT OF DATE. See help cmd.




Could use https://github.com/WireGuard/wgctrl-go + json database.
Root separation by file deposit.

Single executable that behaves much like wg-quick. Not a service, just a
helper.

`dsnet init`
Creates the config file /etc/dsnet.json defining subnet, creating private key,
etc.

`dsnet sync`
Loads peers from JSON file /etc/dsnet.json and brings the interface online. If
interface is already online, synchronises peers by adding/removing. Interface
name in file, dsnet. Runs commands to add routes/forwarding/whatever.

`dsnet down`
Brings the interface down after disassociating all peers.

`dsnet add`
Add a peer by name. Returns a config file as QR code or file as specified. If
public key is specified, private key won't be generated. Editing/removing a
peer can be done by editing the JSON file.

QR code + confirmation prompt on stderr, peer info on stdout.

https://magic-wormhole.readthedocs.io/ (or another "secure" mechanismmechanism
such https://github.com/timvisee/ffsend) could be used to transfer the config
to allow invites.

`dsnet report`
Generates a JSON report listing peers by name, transfer rate, online status, IP
etc. The JSON is intended to be consumed by a hugo template as a data source.
Could also be updated via XHR/websockets.

Report is intended to be generated every minute by cron running as root. The
webserver can then read the file. Location /var/lib/dsnet-report.json
