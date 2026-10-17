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
      lib.mkAndroidEditor = import ./lib/mk-android-editor.nix;
      packages = forSystems (system: {
        default = (editorFor system).jdtls;
        jdtls = (editorFor system).jdtls;
        jdtls-proxy = (editorFor system).proxy;
        java-ide-sources = (editorFor system).sources;
      });
      checks = forSystems (system: {
        proxy = (editorFor system).proxy;
        java-models =
          let
            pkgs = import nixpkgs { inherit system; };
            jdk = pkgs.jdk25_headless;
            gradle = pkgs.gradle_9.override { java = jdk; };
            androidEditor = import ./lib/mk-android-editor.nix { inherit pkgs jdk gradle; };
          in pkgs.runCommand "java-editor-model-regressions" {
            nativeBuildInputs = [ pkgs.python3 jdk gradle ];
            JAVA_HOME = "${jdk}/lib/openjdk";
          } ''
            export HOME="$TMPDIR/home"
            mkdir -p "$HOME"
            python3 ${./tests/gradle-models.py} --helper ${./gradle/ide-sources.gradle} --adapter ${androidEditor.androidInitScript}
            touch "$out"
          '';
      });
      formatter = forSystems (system: nixpkgs.legacyPackages.${system}.nixpkgs-fmt);
    };
}
