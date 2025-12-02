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
      ...
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        inherit (pkgs) lib dockerTools buildEnv cacert;

        version = "0.1.0";

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

        site = pkgs.buildGoModule {
          pname = "jordanmurray-xyz";
          inherit version;
          src = ./.;
          vendorHash = null;

          nativeBuildInputs = [ templ ];

          preBuild = ''
            ${templ}/bin/templ generate
          '';

          ldflags = [
            "-X jordanmurray.xyz/site/version.GitSHA=${self.rev or self.dirtyRev or "unknown"}"
          ];

          postInstall = ''
            mkdir -p $out/share/site
            cp -r static $out/share/site/
            cp -r content $out/share/site/
          '';

          meta = with pkgs.lib; {
            description = "site containing jordanmurray.xyz";
            homepage = "https://jordanmurray.xyz";
            license = licenses.mit;
            mainProgram = "site";
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
            echo "🚀 development environment loaded!"
            echo ""
            echo "`make help`"
          '';
        };

        packages = {
          default = site;

          jordanmurray-xyz = site;

          container = dockerTools.buildImage {
            name = "fueledbyjordan/jordanmurray-xyz";
            tag = version;
            created = "now";

            copyToRoot = buildEnv {
              name = "image-root";
              paths = [
                site
                cacert
              ];
              pathsToLink = [
                "/bin"
                "/share"
                "/etc/ssl/certs"
              ];
            };

            config = {
              Cmd = [ "${lib.getExe site}" ];
              ExposedPorts = {
                "42069/tcp" = { };
              };
              Env = [
                "PORT=42069"
              ];
              User = "10000:10000";
              WorkingDir = "/share/site";
            };
          };
        };
      }
    );
}
