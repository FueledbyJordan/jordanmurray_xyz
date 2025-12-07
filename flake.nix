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
        inherit (pkgs)
          lib
          dockerTools
          buildEnv
          ;

        version = "latest";

        # Fetch external vendor assets with fixed hashes for reproducibility
        daisyuiCss = pkgs.fetchurl {
          url = "https://cdn.jsdelivr.net/npm/daisyui@4.12.14/dist/full.min.css";
          sha256 = "sha256-/fJonUuxkkw6K/ReqJCfynXXQVosFIMt2wuR7WFaNtE=";
        };

        hackFontCss = pkgs.fetchurl {
          url = "https://cdn.jsdelivr.net/npm/hack-font@3/build/web/hack.css";
          sha256 = "sha256-nZuLDiukZ8m5pnOiJdxK1IjXsQ2qbZB1R6x0ddFRyog=";
        };

        hackFontRegular = pkgs.fetchurl {
          url = "https://cdn.jsdelivr.net/npm/hack-font@3/build/web/fonts/hack-regular.woff2";
          sha256 = "sha256-Cw7yVN/Hr8FyUo4xZurOgTmJ4c939Xbdrl9ej7KJfAY=";
        };

        hackFontBold = pkgs.fetchurl {
          url = "https://cdn.jsdelivr.net/npm/hack-font@3/build/web/fonts/hack-bold.woff2";
          sha256 = "sha256-1aZRkOEtHXpOgECF3DzCvIJkug9jNRu4yJBcUB+mmqM=";
        };

        hackFontItalic = pkgs.fetchurl {
          url = "https://cdn.jsdelivr.net/npm/hack-font@3/build/web/fonts/hack-italic.woff2";
          sha256 = "sha256-l/EtxUOQ/SqVY4+MRrmMMlUTbcvgSXD+Q8mw4oPBT70=";
        };

        hackFontBoldItalic = pkgs.fetchurl {
          url = "https://cdn.jsdelivr.net/npm/hack-font@3/build/web/fonts/hack-bolditalic.woff2";
          sha256 = "sha256-1ueYyiSuM+KtmMF2iF9oAO8wNCAad+ArYpM+cNvdIbc=";
        };

        datastarJs = pkgs.fetchurl {
          url = "https://cdn.jsdelivr.net/npm/@sudodevnull/datastar";
          sha256 = "sha256-H4nc9oReY6ZKg9J32E2VkZCAbNXAFXJJv6uypW24+BA=";
        };

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

          nativeBuildInputs = [
            templ
            pkgs.nodePackages.terser
            pkgs.nodePackages.tailwindcss
          ];

          preBuild = ''
            ${templ}/bin/templ generate

            # Create vendor asset directories
            mkdir -p static/vendor/css
            mkdir -p static/vendor/js
            mkdir -p static/vendor/fonts

            # Build Tailwind CSS
            echo "Building Tailwind CSS..."
            NODE_PATH=${pkgs.nodePackages.tailwindcss}/lib/node_modules \
              ${pkgs.nodePackages.tailwindcss}/bin/tailwindcss \
              -i static/css/tailwind.input.css \
              -o static/css/tailwind.css \
              --minify

            # Copy DaisyUI CSS
            cp ${daisyuiCss} static/vendor/css/daisyui.min.css

            # Minify JavaScript assets
            echo "Minifying datastar.js..."
            ${pkgs.nodePackages.terser}/bin/terser ${datastarJs} --compress --mangle -o static/vendor/js/datastar.js

            # Copy Hack font files
            cp ${hackFontRegular} static/vendor/fonts/hack-regular.woff2
            cp ${hackFontBold} static/vendor/fonts/hack-bold.woff2
            cp ${hackFontItalic} static/vendor/fonts/hack-italic.woff2
            cp ${hackFontBoldItalic} static/vendor/fonts/hack-bolditalic.woff2

            # Copy and patch Hack font CSS to use local paths
            cp ${hackFontCss} static/vendor/css/hack.css
            chmod +w static/vendor/css/hack.css
            ${pkgs.gnused}/bin/sed -i 's|fonts/hack-regular\.woff2|/static/vendor/fonts/hack-regular.woff2|g' static/vendor/css/hack.css
            ${pkgs.gnused}/bin/sed -i 's|fonts/hack-bold\.woff2|/static/vendor/fonts/hack-bold.woff2|g' static/vendor/css/hack.css
            ${pkgs.gnused}/bin/sed -i 's|fonts/hack-italic\.woff2|/static/vendor/fonts/hack-italic.woff2|g' static/vendor/css/hack.css
            ${pkgs.gnused}/bin/sed -i 's|fonts/hack-bolditalic\.woff2|/static/vendor/fonts/hack-bolditalic.woff2|g' static/vendor/css/hack.css
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

          site-stripped = site.overrideAttrs (oldAttrs: {
            nativeBuildInputs = oldAttrs.nativeBuildInputs ++ [ pkgs.upx ];
            ldflags = oldAttrs.ldflags ++ [ "-s" "-w" ];
            postInstall = oldAttrs.postInstall + ''
              upx --best --lzma $out/bin/site
            '';
          });

          container = dockerTools.buildImage {
            name = "fueledbyjordan/jordanmurray-xyz";
            tag = version;
            created = "now";

            copyToRoot = buildEnv {
              name = "image-root";
              paths = [
                self.packages.${system}.site-stripped
              ];
              pathsToLink = [
                "/bin"
                "/share"
              ];
            };

            config = {
              Cmd = [ "/bin/site" ];
              ExposedPorts = {
                "9090/tcp" = { };
              };
              Env = [
                "PORT=9090"
              ];
              User = "10000:10000";
              WorkingDir = "/share/site";
            };
          };
        };
      }
    );
}
