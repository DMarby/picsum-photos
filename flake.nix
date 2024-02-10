{
  description = "picsum.photos";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs@{
    self,
    nixpkgs,
    flake-utils,
    ...
  }:
    flake-utils.lib.eachSystem [
      "x86_64-linux"
      "aarch64-linux"
      "aarch64-darwin"
    ] (system:
      let pkgs = nixpkgs.legacyPackages.${system}; in
      {
        packages = rec {
          picsum-photos = pkgs.buildGo122Module {
            name = "picsum-photos";
            src = ./.;
            CGO_ENABLED = 0;
            subPackages = ["cmd/picsum-photos"];
            doCheck = false; # Prevent make test from being ran
            vendorHash = (pkgs.lib.fileContents ./go.mod.sri);
          };
        };

        devShells.default = pkgs.mkShell {
          packages = with pkgs; [
            go_1_22
            gotools
            go-tools
            gopls
            delve
          ];
        };
      }
    );
}
