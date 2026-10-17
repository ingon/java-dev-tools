{
  description = "Reusable Java editor tools for Nix development shells";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

  outputs = { nixpkgs, ... }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "aarch64-darwin" ];
      forSystems = nixpkgs.lib.genAttrs systems;
      mkJavaEditor = import ./lib/mk-java-editor.nix;
      editorFor = system:
        let
          pkgs = import nixpkgs { inherit system; };
          jdk = pkgs.jdk25_headless;
          gradle = pkgs.gradle_9.override { java = jdk; };
        in mkJavaEditor { inherit pkgs jdk gradle; };
    in
    {
      lib.mkJavaEditor = mkJavaEditor;
      packages = forSystems (system: {
        default = (editorFor system).jdtls;
        jdtls = (editorFor system).jdtls;
        jdtls-proxy = (editorFor system).proxy;
        java-ide-sources = (editorFor system).sources;
      });
      checks = forSystems (system: {
        proxy = (editorFor system).proxy;
      });
      formatter = forSystems (system: nixpkgs.legacyPackages.${system}.nixpkgs-fmt);
    };
}
