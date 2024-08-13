
<img src="/etc/logo/banner.svg" alt="Dsnet banner" width="100%">

<a href="https://repology.org/project/dsnet/versions">
    <br />
    <img src="https://repology.org/badge/vertical-allrepos/dsnet.svg" alt="Packaging status" align="right">
</a>

<p>
<a href="https://goreportcard.com/report/github.com/naggie/dsnet"><img src="https://goreportcard.com/badge/github.com/naggie/dsnet" /></a>
<a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-blue.svg" /></a>
<a href="http://godoc.org/github.com/naggie/dsnet"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"/></a>
</p>

<!-- fine line --><h1 align="center"></h1>

<br>
<br>
<br>

Set up a VPN in one minute:


![dsnet add](https://raw.githubusercontent.com/naggie/dsnet/master/etc/init+add.png)

The server peer is listening, and a client peer config has been generated and
added to the server peer:

![wg](https://raw.githubusercontent.com/naggie/dsnet/master/etc/wg2.png)

More client peers can be added with `dsnet add`. They can connect immediately
after! Don't forget to [enable IP forwarding](https://askubuntu.com/questions/311053/how-to-make-ip-forwarding-permanent)
to allow peers to talk to one another.

It works on AMD64 based linux and also ARMv5.

    Usage:
        dsnet [command]

    Available Commands:
      add         Add a new peer + sync
      down        Destroy the interface, run pre/post down
      help        Help about any command
      init        Create /etc/dsnetconfig.json containing default configuration + new keys without loading. Edit to taste.
      regenerate  Regenerate keys and config for peer
      remove      Remove a peer by hostname provided as argument + sync
      report      Generate a JSON status report to stdout
      sync        Update wireguard configuration from /etc/dsnetconfig.json after validating
      up          Create the interface, run pre/post up, sync
      version     Print version

    Flags:
      -h, --help            help for this command
          --output string   config file format: vyatta/wg-quick/nixos (default "wg-quick")

    Use "dsnet [command] --help" for more information about a command.


Quick start (AMD64 linux) -- install wireguard, then, after making sure `/usr/local/bin` is in your path:

    sudo wget https://github.com/naggie/dsnet/releases/latest/download/dsnet-linux-amd64 -O /usr/local/bin/dsnet
    sudo chmod +x /usr/local/bin/dsnet
    sudo dsnet init
    # edit /etc/dsnetconfig.json to taste
    sudo dsnet up
	sudo dsnet add banana > dsnet-banana.conf
	sudo dsnet add apple > dsnet-apple.conf
    # enable IP forwarding to allow peers to talk to one another
    sudo sysctl -w net.ipv4.ip_forward=1   # edit /etc/sysctl.conf to make this persistent across reboots

Copy the generated configuration file to your device and connect!

To send configurations, here are a few suggestions.
- [ffsend](https://github.com/timvisee/ffsend), the most straightforward option;
- [magic wormhole](https://magic-wormhole.readthedocs.io/), a more advanced
  option, where the file never passes through another server;
- [wormhole-william](https://github.com/psanford/wormhole-william), a Go
  implementation of the above.

For the above options, one should transfer the password separately.

A local QR code generator, such as the popular
[qrencode](https://fukuchi.org/works/qrencode/) may also be used to generate a
QR code of the configuration. For instance: `dsnet add | qrencode -t ansiutf8`.
This works because the dsnet prompts are on STDERR and not passed to qrencode.

The peer private key is generated on the server, which is technically not as
secure as generating it on the client peer and then providing the server the
public key; there is provision to specify a public key in the code when adding
a peer to avoid the server generating the private key. The feature will be
added when requested.

Note that named arguments can be specified on the command line as well as
entered by prompt; this allows for unattended usage.

# GUI

Dsnet does not include or require a GUI, however there is now a separate
official monitoring GUI: <https://github.com/botto/dsnet-gui>.

# Configuration overview

The configuration is a single JSON file. Beyond possible initial
customisations, the file is managed entirely by dsnet.

dsnetconfig.json is the only file the server needs to run the VPN. It contains
the server keys, peer public/shared keys and IP settings. **A working version is
automatically generated by `dsnet init` which can be modified as required.**

Currently its location is fixed as all my deployments are for a single network.
I may add a feature to allow setting of the location via environment variable
in the future to support multiple networks on a single host.

Main (automatically generated) configuration example:


    {
        "ExternalHostname": "",
        "ExternalIP": "198.51.100.2",
        "ExternalIP6": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
        "ListenPort": 51820,
        "Domain": "dsnet",
        "InterfaceName": "dsnet",
        "Network": "10.164.236.0/22",
        "Network6": "fd00:7b31:106a:ae00::/64",
        "IP": "10.164.236.1",
        "IP6": "fd00:d631:74ca:7b00:a28:11a1:b821:f013",
        "DNS": "",
        "Networks": [],
        "PrivateKey": "uC+xz3v1mfjWBHepwiCgAmPebZcY+EdhaHAvqX2r7U8=",
        "PostUp": "",
        "PostDown" "",
        "Peers": [
            {
                "Hostname": "test",
                "Owner": "naggie",
                "Description": "Home server",
                "IP": "10.164.236.2",
                "IP6": "fd00:7b31:106a:ae00:44c3:29c3:53b1:a6f9",
                "Added": "2020-05-07T10:04:46.336286992+01:00",
                "Networks": [],
                "PublicKey": "altJeQ/V52JZQrGcA9RiKcpZusYU6zMUJhl7Wbd9rX0=",
                "PresharedKey": "GcUtlze0BMuxo3iVEjpOahKdTf8xVfF8hDW3Ylw5az0="
            }
        ]
    }


See [CONFIG.md](CONFIG.md) for an explanation of each field.


# Report file overview

An example report file, generated by `dsnet report`. Suggested location:
`/var/lib/dsnetreport.json`:

    {
        "ExternalIP": "198.51.100.2",
        "InterfaceName": "dsnet",
        "ListenPort": 51820,
        "Domain": "dsnet",
        "IP": "10.164.236.1",
        "Network": "10.164.236.0/22",
        "DNS": "",
        "PeersOnline": 4,
        "PeersTotal": 13,
        "ReceiveBytes": 32517164,
        "TransmitBytes": 85384984,
        "ReceiveBytesSI": "32.5 MB",
        "TransmitBytesSI": "85.4 MB",
        "Peers": [
            {
                "Hostname": "test",
                "Owner": "naggie",
                "Description": "Home server",
                "Online": false,
                "Dormant": true,
                "Added": "2020-03-12T20:15:42.798800741Z",
                "IP": "10.164.236.2",
                "ExternalIP": "198.51.100.223",
                "Networks": [],
                "Added": "2020-05-07T10:04:46.336286992+01:00",
                "ReceiveBytes": 32517164,
                "TransmitBytes": 85384984,
                "ReceiveBytesSI": "32.5 MB",
                "TransmitBytesSI": "85.4 MB"
            }

            <...>
        ]
    }

Fields mean the same as they do above, or are self explanatory. Note that some
data is converted into human readable formats in addition to machine formats --
this is technically redundant but useful with Hugo shortcodes and other site generators.

The report can be converted, for instance, into a HTML table as below:

![dsnet report table](https://raw.githubusercontent.com/naggie/dsnet/master/etc/report.png)

See
[etc/README.md](https://github.com/naggie/dsnet/blob/master/contrib/report_rendering/README.md)
for hugo and PHP code for rendering a similar table.

# Generating other config files

dsnet currently supports the generation of a `wg-quick` configuration by
default. It can also generate VyOS/Vyatta configuration for EdgeOS/Unifi devices
such as the Edgerouter 4 using the
[wireguard-vyatta](https://github.com/WireGuard/wireguard-vyatta-ubnt) package,
as well as configuration for [NixOS](https://nixos.org), ready to be added to
`configuration.nix` environment definition. [MikroTik RouterOS](https://mikrotik.com/software)
support is also available.

To change the config file format, set the following environment variables:

* `DSNET_OUTPUT=vyatta`
* `DSNET_OUTPUT=wg-quick`
* `DSNET_OUTPUT=nixos`
* `DSNET_OUTPUT=routeros`

Example vyatta output:

    configure
    set interfaces wireguard wg23 address 10.165.52.3/22
    set interfaces wireguard wg23 address fd00:7b31:106a:ae00:f7bb:bf31:201f:60ab/64
    set interfaces wireguard wg23 route-allowed-ips true
    set interfaces wireguard wg23 private-key cAtj1tbjGGmVoxdY78q9Sv0EgNlawbzffGWjajQkLFw=
    set interfaces wireguard wg23 description dsnet

    set interfaces wireguard wg23 peer PjxQM7OwVYvOJfORA1EluLw8CchSu7jLq92YYJi5ohY= endpoint 123.123.123.123:51820
    set interfaces wireguard wg23 peer PjxQM7OwVYvOJfORA1EluLw8CchSu7jLq92YYJi5ohY= persistent-keepalive 25
    set interfaces wireguard wg23 peer PjxQM7OwVYvOJfORA1EluLw8CchSu7jLq92YYJi5ohY= preshared-key w1FtOKoMEdnhsjREtSvpg1CHEKFzFzJWaQYZwaUCV38=
    set interfaces wireguard wg23 peer PjxQM7OwVYvOJfORA1EluLw8CchSu7jLq92YYJi5ohY= allowed-ips 10.165.52.0/22
    set interfaces wireguard wg23 peer PjxQM7OwVYvOJfORA1EluLw8CchSu7jLq92YYJi5ohY= allowed-ips fd00:7b31:106a:ae00::/64
    commit; save

The interface (in this case `wg23`) is deterministically chosen in the range
`wg0-wg999`. This is such that you can use multiple dsnet configurations and
the interface numbers will (probably) be different. The interface number is
arbitrary, so if it is already assigned replace it with a number of your
choice.

Example NixOS output:

    networking.wireguard.interfaces = {
      dsnet = {
        ips = [
          "10.9.8.2/22"
          "fd00:80f8:af4a:4700:aaaa:bbbb:cccc:88ad/64"
          ];
        privateKey = "2PvML6bsmTCK+cBxpV9SfF261fsH6gICixtppfG6KFc=";
        peers = [
          {
            publicKey = "zCDo5yn7Muy3mPBXtarwm5S7JjNKM0IdIdGqoreWmSA=";
            presharedKey = "5Fa8Zc8gIkpfBPJUJn5OEVuE00iqmXnS34v4evv1MUM=";
            allowedIPs = [
              "10.56.72.0/22"
              "fd00:80f8:af4a:4700::/64"
              ];
            endpoint = "123.123.123.123:51820";
            persistentKeepalive = 25;
          }
        ];
      };
    };

Example MikroTik RouterOS output:

    /interface wireguard
    add name=wg0 private-key="CDWdi0IcMZgla1hCYI41JejjuFaPCle+vPBxvX5OvVE=";
    /interface list member
    add interface=wg0 list=LAN
    /ip address
    add address=10.55.148.2/22 interface=wg0
    /ipv6 address
    add address=fd00:1965:946d:5000:5a88:878d:dc0:c777/64 advertise=no eui-64=no no-dad=no interface=wg0
    /interface wireguard peers
    add interface=wg0 \
        public-key="iE7dleTu34JOCC4A8xdIZcnbNE+aoji8i1JpP+gdt0M=" \
        preshared-key="Ch0BdZ6Um29D34awlWBSNa+cz1wGOUuHshjYIyqKxGU=" \
        endpoint-address=198.51.100.73 \
        endpoint-port=51820 \
        persistent-keepalive=25s \
        allowed-address=10.55.148.0/22,fd00:1965:946d:5000::/64,192.168.10.0/24,fe80::1/64

# FAQ

> Does dsnet support IPv6?

Yes! By default since version 0.2, a random ULA subnet is generated with a 0
subnet ID. Peers are allocated random addresses when added. Existing IPv4
configs will not be updated -- add a `Network6` subnet to the existing config
to allocate addresses to new peers.

Like IPv4, it's up to you if you want to provide NAT IPv6 access to the
internet; alternatively (and preferably) you can allocate a a real IPv6 subnet
such that all peers have a real globally routeable IPv6 address.

Upon initialisation, the server IPv4 and IPv6 external IP addresses are
discovered on a best-effort basis. Clients will have configuration configured
for the server IPv4 preferentially. If not IPv4 is configured, IPv6 is used;
this is to give the best chance of the VPN working regardless of the dodgy
network you're on.

> Is dsnet production ready?

Absolutely, it's just a configuration generator so your VPN does not depend on
dsnet after adding peers. I use it in production at 2 companies so far.

Note that before version 1.0, the config file schema may change. Changes will
be made clear in release notes.

> Client private keys are generated on the server. Can I avoid this?

Allowing generation of the pub/priv keypair on the client is not yet supported,
but will be soon as provision exists within the code base. Note that whilst
client peer private keys are generated on the server, they are never stored.


> How do I get dsnet to bring the (server) interface up on startup?

Assuming you're running a systemd powered linux distribution (most of them are):

1. Copy
   [etc/dsnet.service](https://github.com/naggie/dsnet/blob/master/etc/dsnet.service)
   to `/etc/systemd/system/`
2. Run `sudo systemctl daemon-reload` to get systemd to see it
3. Then run `sudo systemctl enable dsnet` to enable it at boot

> How can I generate the report periodically?

Either with cron or a systemd timer. Cron is easiest:

    echo '* * * * * root /usr/local/bin/dsnet report | sudo tee /etc/cron.d/dsnetreport'

Note that whilst report generation requires root, consuming the report does not
as it's just a world-readable file. This is important for web interfaces that
need to be secure.

This is also why dsnet loads its configuration from a file -- it's possible to
set permissions such that dsnet synchronises the config generated by a non-root
user. Combined with a periodic `dsnet sync` like above, it's possible to build
a secure web interface that does not require root. A web interface is currently
being created by a friend; it will not be part of dstask, rather a separate
project.

----

The dsnet logo was kindly designed by [@mirorauhala](https://github.com/mirorauhala).
