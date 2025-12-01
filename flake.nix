{
  description = "jordanmurray.xyz";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        templ = pkgs.buildGoModule rec {
          pname = "templ";
          version = "0.3.960";

          src = pkgs.fetchFromGitHub {
            owner = "a-h";
            repo = "templ";
            rev = "v${version}";
            hash = "sha256-GCbqaRC9KipGdGfgnGjJu04/rJlg+2lgi2vluP05EV4=";
          };

          vendorHash = "sha256-pVZjZCXT/xhBCMyZdR7kEmB9jqhTwRISFp63bQf6w5A=";

          subPackages = [ "cmd/templ" ];

          meta = with pkgs.lib; {
            description = "A language for writing HTML user interfaces in Go";
            homepage = "https://templ.guide/";
            license = licenses.mit;
            mainProgram = "templ";
          };
        };
      in
      {
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            templ
            gnumake
            air
          ];

          shellHook = ''
            echo "🚀 Blog development environment loaded!"
            echo ""
            echo "Available commands:"
            echo "  make install  - Install dependencies"
            echo "  make run      - Generate templates and start server"
            echo "  make watch    - Watch templates for changes"
            echo "  make dev      - Run with hot reload (using air)"
            echo "  make build    - Build production binary"
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "jordanmurray-xyz";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;

          nativeBuildInputs = [ templ ];

          preBuild = ''
            ${templ}/bin/templ generate
          '';

          ldflags = [
            "-X jordanmurray.xyz/blog/version.NixSHA=${self.rev or self.dirtyRev or "dev"}"
            "-X jordanmurray.xyz/blog/version.GitSHA=${self.rev or self.dirtyRev or "unknown"}"
          ];

          meta = with pkgs.lib; {
            description = "site containing jordanmurray.xyz";
            homepage = "https://jordanmurray.xyz";
            license = licenses.mit;
            mainProgram = "jordanmurray_xyz";
          };
        };
      }
    );
}
