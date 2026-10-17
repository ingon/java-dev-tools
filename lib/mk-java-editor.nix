{ pkgs, jdk, gradle ? null }:
let
  proxy = pkgs.callPackage ../pkgs/jdtls-proxy.nix { };
  server = pkgs.jdt-language-server.override { inherit jdk; };
  jdtls = pkgs.writeShellScriptBin "jdtls" ''
    set -eu
    export JAVA_HOME=${pkgs.lib.escapeShellArg "${jdk}/lib/openjdk"}
    ${pkgs.lib.optionalString (gradle != null) ''
      export GRADLE_HOME=${pkgs.lib.escapeShellArg "${gradle}/libexec/gradle"}
    ''}
    if [ -n "''${JDTLS_WORKSPACE:-}" ]; then
      workspace="$JDTLS_WORKSPACE"
    else
      project_root="$(${pkgs.gitMinimal}/bin/git rev-parse --show-toplevel 2>/dev/null || pwd -P)"
      project_key="$(printf '%s' "$project_root" | ${pkgs.coreutils}/bin/sha256sum)"
      project_key="''${project_key%% *}"
      workspace="''${XDG_CACHE_HOME:-$HOME/.cache}/jdtls/$project_key"
    fi
    ${pkgs.coreutils}/bin/mkdir -p "$workspace"
    # The proxy resolves its child by name. Put the real server before this launcher.
    export PATH="${server}/bin:$PATH"
    exec ${proxy}/bin/jdtls-wrapper -data "$workspace" "$@"
  '';
  sources = if gradle == null then null else
    pkgs.writeShellScriptBin "java-ide-sources" ''
      set -eu
      export JAVA_HOME=${pkgs.lib.escapeShellArg "${jdk}/lib/openjdk"}
      exec ${gradle}/bin/gradle --no-configuration-cache \
        --init-script ${../gradle/ide-sources.gradle} ideSources "$@"
    '';
in
{
  inherit jdtls proxy sources;
  packages = [ jdtls ] ++ pkgs.lib.optional (sources != null) sources;
}
