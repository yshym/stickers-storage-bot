{
  description = "Stickers storage bot";

  inputs = {
    nixpkgs.url = "github:yshym/nixpkgs/release-21.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        devShell = with pkgs;
          mkShell {
            nativeBuildInputs = [ gnumake ];
          };
      });
}
