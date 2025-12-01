{
  description = "Jordan Murray's Blog - Go, Templ, Datastar, DaisyUI";

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
            air # Optional: for hot reloading
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
            echo ""
            echo "Tech stack: Go + Templ + Datastar + DaisyUI"
          '';
        };

        packages.default = pkgs.buildGoModule {
          pname = "jordanmurray-blog";
          version = "0.1.0";
          src = ./.;
          vendorHash = null;

          nativeBuildInputs = [ templ ];

          preBuild = ''
            ${templ}/bin/templ generate
          '';

          ldflags = [
            "-X jordanmurray.xyz/blog/version.Version=${self.rev or self.dirtyRev or "dev"}"
          ];

          meta = with pkgs.lib; {
            description = "Jordan Murray's blog built with Go, Templ, Datastar, and DaisyUI";
            homepage = "https://jordanmurray.xyz";
            license = licenses.mit;
            mainProgram = "blog";
          };
        };
      }
    );
}
