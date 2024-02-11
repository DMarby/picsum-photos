{
  description = "picsum.photos";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
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
      let pkgs = nixpkgs.legacyPackages.${system}; in {
        packages = rec {
          default = everything;

          everything = pkgs.symlinkJoin {
            name = "picsum-photos-composite";
            paths = [
              picsum-photos
              image-service
            ];
          };

          picsum-photos = pkgs.buildGo122Module {
            name = "picsum-photos";
            src = ./.;
            CGO_ENABLED = 0;
            subPackages = ["cmd/picsum-photos"];
            doCheck = false; # Prevent make test from being ran
            vendorHash = (pkgs.lib.fileContents ./go.mod.sri);
            nativeBuildInputs = with pkgs; [
              tailwindcss
            ];
            preBuild = ''
              go generate ./...
            '';
          };

          image-service = pkgs.buildGo122Module {
            name = "image-service";
            src = ./.;
            subPackages = ["cmd/image-service"];
            doCheck = false; # Prevent make test from being ran
            vendorHash = (pkgs.lib.fileContents ./go.mod.sri);
            nativeBuildInputs = with pkgs; [
              pkg-config
            ];
            buildInputs = with pkgs; [
              vips
            ];
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
    ) // {
      nixosModules.default = { config, lib, pkgs, ... }:
        with lib;
        let cfg = config.picsum-photos.services;
        in {
          options.picsum-photos.services = {
            picsum-photos = {
              enable = mkEnableOption "Enable the picsum-photos service";

              logLevel = mkOption {
                type = with types; enum [ "debug" "info" "warn" "error" "dpanic" "panic" "fatal" ];
                example = "debug";
                default = "info";
                description = "log level";
              };

              domain = mkOption {
                type = types.str;
                description = "Domain to listen to";
              };

              sockPath = mkOption rec {
                type = types.path;
                default = "/run/picsum-photos/picsum-photos.sock";
                example = default;
                description = "Unix domain socket to listen on";
              };

              environmentFile = mkOption {
                type = types.path;
                description = "Environment file";
              };

              databaseFilePath = mkOption rec {
                type = types.path;
                default = "/var/lib/picsum-photos/image-manifest.json";
                example = default;
                description = "Image database file path";
              };
            };

            image-service = {
              enable = mkEnableOption "Enable the image-service service";

              logLevel = mkOption {
                type = with types; enum [ "debug" "info" "warn" "error" "dpanic" "panic" "fatal" ];
                example = "debug";
                default = "info";
                description = "log level";
              };

              workers = mkOption rec {
                type = types.number;
                default = 16;
                example = default;
                description = "worker queue concurrency";
              };

              domain = mkOption {
                type = types.str;
                description = "Domain to listen to";
              };

              sockPath = mkOption rec {
                type = types.path;
                default = "/run/image-service/image-service.sock";
                example = default;
                description = "Unix domain socket to listen on";
              };

              environmentFile = mkOption {
                type = types.path;
                description = "Environment file";
              };

              storagePath = mkOption rec {
                type = types.path;
                default = "/var/lib/image-service";
                example = default;
                description = "Storage path";
              };
            };
          };

          config = mkMerge([
            (mkIf cfg.picsum-photos.enable {
              users.groups.picsum-photos = {};

              users.users.picsum-photos = {
                createHome = true;
                isSystemUser = true;
                group = "picsum-photos";
                home = "/var/lib/picsum-photos";
              };

              systemd.services.picsum-photos = {
                description = "picsum-photos";
                wantedBy = [ "multi-user.target" ];

                script = ''
                  exec ${self.packages.${pkgs.system}.picsum-photos}/bin/picsum-photos \
                    -log-level=${cfg.picsum-photos.logLevel} \
                    -listen=${cfg.picsum-photos.sockPath} \
                    -database-file-path=${cfg.picsum-photos.storagePath}
                '';

                serviceConfig = {
                  EnvironmentFile = cfg.picsum-photos.environmentFile;
                  User = "picsum-photos";
                  Group = "picsum-photos";
                  Restart = "always";
                  RestartSec = "30s";
                  WorkingDirectory = "/var/lib/picsum-photos";
                  RuntimeDirectory = "picsum-photos";
                  RuntimeDirectoryMode = "0770";
                  UMask = "007";
                };
              };

              services.nginx.virtualHosts."${cfg.picsum-photos.domain}" = {
                locations."/" = {
                  proxyPass = "http://unix:${cfg.picsum-photos.sockPath}";
                };
              };
            })

            (mkIf cfg.image-service.enable {
              users.groups.image-service = {};

              users.users.image-service = {
                createHome = true;
                isSystemUser = true;
                group = "image-service";
                home = "/var/lib/image-service";
              };

              systemd.services.image-service = {
                description = "image-service";
                wantedBy = [ "multi-user.target" ];

                script = ''
                  exec ${self.packages.${pkgs.system}.image-service}/bin/image-service \
                    -log-level=${cfg.image-service.logLevel} \
                    -listen=${cfg.image-service.sockPath} \
                    -storage-path=${cfg.image-service.storagePath} \
                    -workers=${cfg.image-service.workers}
                '';

                serviceConfig = {
                  EnvironmentFile = cfg.image-service.environmentFile;
                  User = "image-service";
                  Group = "image-service";
                  Restart = "always";
                  RestartSec = "30s";
                  WorkingDirectory = "/var/lib/image-service";
                  RuntimeDirectory = "image-service";
                  RuntimeDirectoryMode = "0770";
                  UMask = "007";
                };
              };

              services.nginx.virtualHosts."${cfg.image-service.domain}" = {
                locations."/" = {
                  proxyPass = "http://unix:${cfg.image-service.sockPath}";
                };
              };
            })
          ]);
        };
    };
}
