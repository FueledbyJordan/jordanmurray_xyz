package version

// Version will be set at build time via -ldflags (Nix flake hash or "dev")
var Version = "dev"

// GitSHA will be set at build time via -ldflags
var GitSHA = "unknown"
