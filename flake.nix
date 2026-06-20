{
  description = "Copy Immich library photos to an NFS share, exactly once";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.11";
  inputs.nixpkgs-unstable.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    {
      self,
      nixpkgs,
      nixpkgs-unstable,
    }:
    let
      supportedSystems = [ "x86_64-linux" ];
      forEachSupportedSystem =
        f:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            pkgs = import nixpkgs { inherit system; };
            pkgs-unstable = import nixpkgs-unstable { inherit system; };
            inherit system;
          }
        );
    in
    {
      devShells = forEachSupportedSystem (
        { pkgs, pkgs-unstable, ... }:
        {
          default = pkgs.mkShell {
            packages = with pkgs; [
              just
              pkgs-unstable.go
              pkgs-unstable.golangci-lint
            ];
          };
        }
      );

      # dev shell only; deployment is the Docker image
    };
}
