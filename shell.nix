{ pkgs ? import <nixpkgs> { } }:

pkgs.mkShell {
  buildInputs = with pkgs; [
    go

    gnumake
    upx

    wireguard-tools

    gopls
    gotools # goimports, godoc, etc.
    go-tools # staticcheck
    delve # debugger
  ];

  CGO_ENABLED = "0";

  shellHook = ''
    echo "dsnet dev shell"
    echo "  go $(go version | awk '{print $3}')"
    echo ""
    echo "  make quick   — fast build"
    echo "  make build   — optimised build (upx)"
    echo "  go test ./...— run tests"
  '';
}
