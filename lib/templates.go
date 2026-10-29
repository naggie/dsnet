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
PersistentKeepalive={{ .Server.PersistentKeepalive }}
{{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
AllowedIPs={{ .Server.Network.IPNet.IP }}/{{ .CidrSize }}
{{ end -}}
{{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
AllowedIPs={{ .Server.Network6.IPNet.IP }}/{{ .CidrSize6  }}
{{ end -}}
{{ range .Server.Networks -}}
AllowedIPs={{ . }}
{{ end -}}
`

const vyattaPeerConf = `configure
{{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
set interfaces wireguard wg0 address {{ .Peer.IP }}/{{ .CidrSize }}
{{ end -}}
{{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
set interfaces wireguard wg0 address {{ .Peer.IP6 }}/{{ .CidrSize6 }}
{{ end -}}
set interfaces wireguard wg0 route-allowed-ips true
set interfaces wireguard wg0 private-key {{ .Peer.PrivateKey.Key }}
set interfaces wireguard wg0 description {{ .Server.InterfaceName }}
{{- if .Server.DNS }}
#set service dns forwarding name-server {{ .Server.DNS }}
{{ end }}

set interfaces wireguard wg0 peer {{ .Server.PrivateKey.PublicKey.Key }} endpoint {{ .Endpoint }}:{{ .Server.ListenPort }}
set interfaces wireguard wg0 peer {{ .Server.PrivateKey.PublicKey.Key }} persistent-keepalive {{ .Server.PersistentKeepalive }}
set interfaces wireguard wg0 peer {{ .Server.PrivateKey.PublicKey.Key }} preshared-key {{ .Peer.PresharedKey.Key }}
{{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
set interfaces wireguard wg0 peer {{ .Server.PrivateKey.PublicKey.Key }} allowed-ips {{ .Server.Network.IPNet.IP }}/{{ .CidrSize }}
{{ end -}}
{{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
set interfaces wireguard wg0 peer {{ .Server.PrivateKey.PublicKey.Key }} allowed-ips {{ .Server.Network6.IPNet.IP }}/{{ .CidrSize6  }}
{{ end -}}
{{ range .Server.Networks -}}
set interfaces wireguard wg0 peer {{ .Server.PrivateKey.PublicKey.Key }} allowed-ips {{ . }}
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
    # Inline keys end up in the world-readable /nix/store. Insecure: prefer
    # privateKeyFile and presharedKeyFile from a secrets manager such as
    # sops-nix or agenix.
    privateKey = "{{ .Peer.PrivateKey.Key }}";
    # Re-resolve the endpoint periodically so a DNS failure at boot or a
    # dynamic IP change heals itself. Interface level, not per peer: the
    # per-peer option fails evaluation with the networkd backend.
    dynamicEndpointRefreshSeconds = 300;
    peers = [
      {{ "{" }}
        publicKey = "{{ .Server.PrivateKey.PublicKey.Key }}";
        presharedKey = "{{ .Peer.PresharedKey.Key }}";
        allowedIPs = [
          {{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
          "{{ .Server.Network.IPNet.IP }}/{{ .CidrSize }}"
          {{ end -}}
          {{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
          "{{ .Server.Network6.IPNet.IP }}/{{ .CidrSize6  }}"
          {{ end -}}
          {{ range .Server.Networks -}}
          "{{ . }}"
          {{ end -}}
        ];
        endpoint = "{{ .Endpoint }}:{{ .Server.ListenPort }}";
        persistentKeepalive = {{ .Server.PersistentKeepalive }};
      {{ "}" }}
    ];
  {{ "};" }}
{{ "};" }}
`

const routerosPeerConf = `/interface wireguard
add name=wg0 private-key="{{ .Peer.PrivateKey.Key }}";
/interface list member
add interface=wg0 list=LAN
/ip address
{{ if gt (.Server.Network.IPNet.IP | len) 0 -}}
add address={{ .Peer.IP }}/{{ .CidrSize }} interface=wg0
{{ end -}}
/ipv6 address
{{ if gt (.Server.Network6.IPNet.IP | len) 0 -}}
add address={{ .Peer.IP6 }}/{{ .CidrSize6 }} advertise=no interface=wg0
{{ end -}}
/interface wireguard peers
{{/* MikroTik RouterOS does not like trailing commas in arrays */ -}}
{{ $first := true -}}
add interface=wg0 \
    public-key="{{ .Server.PrivateKey.PublicKey.Key }}" \
    preshared-key="{{ .Peer.PresharedKey.Key }}" \
    endpoint-address={{ .Endpoint }} \
    endpoint-port={{ .Server.ListenPort }} \
    persistent-keepalive={{ .Server.PersistentKeepalive }}s \
    allowed-address=
        {{- if gt (.Server.Network.IPNet.IP | len) 0 }}
            {{- if $first}}{{$first = false}}{{else}},{{end}}
            {{- .Server.Network.IPNet.IP }}/{{ .CidrSize }}
        {{- end }}
        {{- if gt (.Server.Network6.IPNet.IP | len) 0 }}
            {{- if $first}}{{$first = false}}{{else}},{{end}}
            {{- .Server.Network6.IPNet.IP }}/{{ .CidrSize6 }}
        {{- end }}
        {{- range .Server.Networks }}
            {{- if $first}}{{$first = false}}{{else}},{{end}}
            {{- . }}
        {{- end }}
`
