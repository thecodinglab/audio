{
  description = "A very basic flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs =
    { self, nixpkgs }:
    let
      systems = [
        "aarch64-linux"
        "x86_64-linux"
        "aarch64-darwin"
        "x86_64-darwin"
      ];

      forAllSystems = nixpkgs.lib.genAttrs systems;
    in
    {
      devShells = forAllSystems (
        system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          default = pkgs.mkShell {
            packages =
              [
                pkgs.pkg-config
                pkgs.libopus
                pkgs.libogg
                pkgs.opusfile
              ]
              ++ pkgs.lib.optionals pkgs.stdenv.isLinux [
                pkgs.gdb
                pkgs.pipewire
              ];
          };
        }
      );

    };
}
