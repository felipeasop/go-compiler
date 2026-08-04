{
  description = "Compilador para um subconjunto da linguagem Go, escrito em Go";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];
      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f nixpkgs.legacyPackages.${system});
    in
    {
      # nix develop → shell com Go e ferramentas
      devShells = forAllSystems (pkgs: {
        default = pkgs.mkShell {
          packages = [ pkgs.go pkgs.gopls pkgs.gotools ];
          shellHook = ''
            echo "go-compiler: $(go version)"
            echo "  go run . testdata/programs/fatorial.go   # compila e executa"
            echo "  go test ./...                         # roda os testes"
          '';
        };
      });

      # nix build → ./result/bin/go-compiler (roda 'go test' durante o build)
      packages = forAllSystems (pkgs: {
        default = pkgs.buildGoModule {
          pname = "go-compiler";
          version = "1.0.0";
          src = ./.;
          vendorHash = null; # sem dependências externas
        };
      });

      # nix run . -- testdata/programs/fatorial.go
      apps = forAllSystems (pkgs: {
        default = {
          type = "app";
          program = "${self.packages.${pkgs.system}.default}/bin/go-compiler";
        };
      });

      # nix flake check → compila e roda todos os testes
      checks = forAllSystems (pkgs: {
        default = self.packages.${pkgs.system}.default;
      });
    };
}
