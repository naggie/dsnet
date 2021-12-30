package lib

const wgQuickPeerConf = `[Interface]
{{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
Address={{ .Peer.IP }}/{{ .CidrSize }}
{{ end -}}
{{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
Address={{ .Peer.IP6 }}/{{ .CidrSize6 }}
{{ end -}}
PrivateKey={{ .Peer.PrivateKey.Key }}
{{- if .Server.DNS }}
DNS={{ .Server.DNS }}
{{ end }}

[Peer]
PublicKey={{ .Server.PrivateKey.PublicKey.Key }}
PresharedKey={{ .Peer.PresharedKey.Key }}
Endpoint={{ .Endpoint }}:{{ .Server.ListenPort }}
PersistentKeepalive={{ .Keepalive }}
{{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
AllowedIPs={{ .Server.Network }}
{{ end -}}
{{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
AllowedIPs={{ .Server.Network6 }}
{{ end -}}
{{ range .Server.Networks -}}
AllowedIPs={{ . }}
{{ end -}}
`

// TODO use random wg0-wg999 to hopefully avoid conflict by default?
const vyattaPeerConf = `configure
{{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
set interfaces wireguard {{ .Wgif }} address {{ .Peer.IP }}/{{ .CidrSize }}
{{ end -}}
{{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
set interfaces wireguard {{ .Wgif }} address {{ .Peer.IP6 }}/{{ .CidrSize6 }}
{{ end -}}
set interfaces wireguard {{ .Wgif }} route-allowed-ips true
set interfaces wireguard {{ .Wgif }} private-key {{ .Peer.PrivateKey.Key }}
set interfaces wireguard {{ .Wgif }} description {{ .Server.InterfaceName }}
{{- if .Server.DNS }}
#set service dns forwarding name-server {{ .Server.DNS }}
{{ end }}

set interfaces wireguard {{ .Wgif }} peer {{ .Server.PrivateKey.PublicKey.Key }} endpoint {{ .Endpoint }}:{{ .Server.ListenPort }}
set interfaces wireguard {{ .Wgif }} peer {{ .Server.PrivateKey.PublicKey.Key }} persistent-keepalive {{ .Keepalive }}
set interfaces wireguard {{ .Wgif }} peer {{ .Server.PrivateKey.PublicKey.Key }} preshared-key {{ .Peer.PresharedKey.Key }}
{{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
set interfaces wireguard {{ .Wgif }} peer {{ .Server.PrivateKey.PublicKey.Key }} allowed-ips {{ .Server.Network }}
{{ end -}}
{{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
set interfaces wireguard {{ .Wgif }} peer {{ .Server.PrivateKey.PublicKey.Key }} allowed-ips {{ .Server.Network6 }}
{{ end -}}
{{ range .Server.Networks -}}
set interfaces wireguard {{ .Wgif }} peer {{ .Server.PrivateKey.PublicKey.Key }} allowed-ips {{ . }}
{{ end -}}
commit; save
`

const nixosPeerConf = `networking.wireguard.interfaces = {{ "{" }}
  dsnet = {{ "{" }}
    ips = [
      {{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
      "{{ .Peer.IP }}/{{ .CidrSize }}"
      {{ end -}}
      {{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
      "{{ .Peer.IP6 }}/{{ .CidrSize6 }}"
      {{ end -}}
    ];
    privateKey = "{{ .Peer.PrivateKey.Key }}";
    {{- if .Server.DNS }}
    dns = [ "{{ .Server.DNS }}" ];
    {{ end }}
    peers = [
      {{ "{" }}
        publicKey = "{{ .Server.PrivateKey.PublicKey.Key }}";
        presharedKey = "{{ .Peer.PresharedKey.Key }}";
        allowedIPs = [
          {{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
          "{{ .Server.Network }}"
          {{ end -}}
          {{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
          "{{ .Server.Network6 }}"
          {{ end -}}
        ];
        endpoint = "{{ .Endpoint }}:{{ .Server.ListenPort }}";
        persistentKeepalive = {{ .Keepalive }};
      {{ "}" }}
    ];
  {{ "};" }}
{{ "};" }}
`
