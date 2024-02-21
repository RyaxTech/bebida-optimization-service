{
  description = "Bebida Shaker";

  # Nixpkgs / NixOS version to use.
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    let
      version = "0.0.2";
      systems = [
        "x86_64-linux"
        # "aarch64-linux"
      ];
      inherit (flake-utils.lib) eachSystem filterPackages;
      bebidaShakerPackage =
        { pkgs }:
        pkgs.buildGoModule {
          pname = "bebida-shaker";
          inherit version;
          src = ./.;
          #src = fetchFromGitHub {
          #  owner = "RyaxTech";
          #  repo = "bebida-shaker";
          #  rev = "main";
          #  sha256 = pkgs.lib.fakeSha256;
          #}

          checkPhase = "";
          # vendorHash = null;
          # vendorHash = pkgs.lib.fakeHash;
          vendorHash = "sha256-n+Pe2nVWlwDLPbzaWTSYtMyYLzMpC1H+oilg7YJhftI=";
        };
      bebidaShakerModule = { config, lib, pkgs, ... }:
        let
          cfg = config.services.bebida-shaker;
        in
        with lib; {
          # interface
          options.services.bebida-shaker = {
            enable = mkEnableOption (lib.mdDoc "bebida-shaker");

            package = mkOption {
              type = types.package;
              default = bebidaShakerPackage { inherit pkgs; };
              defaultText = literalExpression "pkgs.bebida-shaker";
              description = lib.mdDoc "Package that should be used for Bebida Shaker";
            };
            environmentFile = mkOption {
              type = types.nullOr types.path;
              description = lib.mdDoc ''
                File path containing environment variables for configuring the Bebida Shaker service in the format of an EnvironmentFile. See systemd.exec(5).
              '';
              default = null;
            };
          };

          # Implementation
          config = mkIf cfg.enable {

            environment.systemPackages = [ config.services.bebida-shaker.package ];

            systemd.services.bebida-shaker = {
              description = "BeBiDa Shaker service";
              after = [ "firewall.service" "network-online.target" ];
              wants = [ "firewall.service" "network-online.target" ];
              wantedBy = [ "multi-user.target" ];
              serviceConfig = {
                Type = "exec";
                KillMode = "process";
                Delegate = "yes";
                Restart = "always";
                RestartSec = "5s";
                EnvironmentFile = cfg.environmentFile;
                ExecStart = "${cfg.package}/bin/bebida-shaker";
              };
            };
          };
        };
    in
    eachSystem systems
      (system:
        let
          pkgs = import nixpkgs { inherit system; };
          bebidaShaker = bebidaShakerPackage { inherit pkgs; };
        in
        {
          packages = {
            bebida-shaker = bebidaShaker;
            default = bebidaShaker;
          };
          formatter = pkgs.nixpkgs-fmt;
        }
      ) // {
      nixosModules = {
        bebida-shaker = bebidaShakerModule;
        default = bebidaShakerModule;
      };
    };
}
