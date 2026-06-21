{
  description = "File-based task queue for AI agent loops";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = self.shortRev or self.dirtyShortRev or "dev";
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "fl";
          inherit version;
          src = ./apps/fl;
          vendorHash = "sha256-8BhEEo+fd12bZDcvGDR2Xs8LyFoBu8YNwGFxhnji5fo=";
          subPackages = [ "cmd/fl" ];
          ldflags = [ "-s" "-w" "-X main.version=${version}" ];
          meta = {
            description = "File-based task queue for AI agent loops";
            homepage = "https://github.com/jeprecated/frontloop";
            license = pkgs.lib.licenses.mit;
            mainProgram = "fl";
          };
        };
      }
    ) // {
      overlays.default = final: prev: {
        fl = self.packages.${final.system}.default;
      };
    };
}
