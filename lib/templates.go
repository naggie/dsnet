package lib

import _ "embed"

//go:embed templates/wg-quick.conf.tmpl
var wgQuickPeerConf string

//go:embed templates/vyatta.tmpl
var vyattaPeerConf string

//go:embed templates/nixos.nix.tmpl
var nixosPeerConf string

//go:embed templates/routeros.rsc.tmpl
var routerosPeerConf string
