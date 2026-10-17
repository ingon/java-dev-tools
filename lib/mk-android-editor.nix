{ pkgs, jdk, gradle, serverJdk ? pkgs.jdk21_headless
, variant ? "debug", unitTests ? true, aapt2 ? null, downloadSources ? false
}:
assert builtins.match "[A-Za-z][A-Za-z0-9]*" variant != null;
let
  androidInitScript = pkgs.writeText "android-ide.init.gradle"
    (builtins.replaceStrings
      [ "@variant@" "@unitTests@" "@aapt2@" "@downloadSources@" ]
      (map builtins.toJSON [ variant unitTests aapt2 downloadSources ])
      (builtins.readFile ../gradle/android-ide.init.gradle));
  editor = import ./mk-java-editor.nix {
    inherit pkgs jdk gradle serverJdk;
    initializationSettings = {
      "java.import.gradle.arguments" = [ "--init-script" (toString androidInitScript) ];
      "java.configuration.updateBuildConfiguration" = "automatic";
    };
  };
in editor // {
  inherit androidInitScript;
  # The standard source helper selects Java-plugin projects, not Android ones.
  sources = null;
  packages = [ editor.jdtls ];
}
