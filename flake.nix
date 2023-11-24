{
  description = "Bebida Shaker";

  # Nixpkgs / NixOS version to use.
  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.05";
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
          # In 'nix develop', we don't need a copy of the source tree
          # in the Nix store.
          src = ./.;

          checkPhase = "";

          # This hash locks the dependencies of this package. It is
          # necessary because of how Go requires network access to resolve
          # VCS.  See https://www.tweag.io/blog/2021-03-04-gomod2nix/ for
          # details. Normally one can build with a fake sha256 and rely on native Go
          # mechanisms to tell you what the hash should be or determine what
          # it should be "out-of-band" with other tooling (eg. gomod2nix).
          # To begin with it is recommended to set this, but one must
          # remeber to bump this hash when your dependencies change.
          #vendorSha256 = pkgs.lib.fakeSha256;

          vendorSha256 = "sha256-F9843vH95xAsvtEsnO6LiSu6MjAg0Ax55l02U/zoFCA=";
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
