# Java development tools for Nix

JDTLS source navigation for Helix and a Gradle dependency-source helper, reusable
across Java projects supported by JDTLS. Use your project's Nix packages, JDK and
optional Gradle; no project-specific editor configuration or copied scripts are
required.

For an existing repository, follow the [integration guide](INTEGRATION.md). It
gives humans and agents the changes to `flake.nix`, optional `build.gradle.kts`
output layout, `.gitignore`, source verification, and the validation handoff.

The `jdtls` command bridges JDTLS's virtual source documents to local files that
Helix can open. It enables the source capability during initialization, so
Helix's built-in Java language configuration works with JDK and library sources.
Source archives must be available and accepted by the project's dependency policy.

Android applications and libraries have a separate opt-in
[`mkAndroidEditor` integration](#android-gradle-projects). Ordinary Java projects
continue to use `mkJavaEditor` unchanged.

## Add to a development shell

Add the published flake input:

```nix
inputs.java-dev-tools = {
  url = "github:ingon/java-dev-tools";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

Include `java-dev-tools` in your flake's `outputs` arguments. With your existing
`pkgs`, `jdk` and `gradle`, compose the shell:

```nix
let
  javaEditor = java-dev-tools.lib.mkJavaEditor { inherit pkgs jdk gradle; };
in {
  devShells.default = pkgs.mkShell (javaEditor.shellEnv // {
    JAVA_HOME = "${jdk}/lib/openjdk";
    packages = [ jdk gradle ] ++ javaEditor.packages;
  });
}
```

The library uses the caller's packages and runs JDTLS on the supplied JDK. Choose
a JDK supported by that JDTLS version; the project's build configuration controls
its compilation target. Project dependencies, build settings and publication
configuration stay in the project. The consuming flake's lockfile pins this input.
If the build JDK is too old to run JDTLS, supply `serverJdk` separately. It defaults
to `jdk`; the launcher retains `JAVA_HOME` for the build JDK in either case.

For Maven or another supported non-Gradle setup, omit `gradle`:

```nix
javaEditor = java-dev-tools.lib.mkJavaEditor { inherit pkgs jdk; };
```

Install/configure your build tool separately. This variant provides `jdtls`
without the Gradle source helper or Gradle shell settings.

`mkJavaEditor` returns:

| Attribute | Purpose |
| --- | --- |
| `jdtls` | Launcher with source navigation and a workspace per checkout. |
| `proxy` | The underlying virtual-source bridge package. |
| `sources` | `java-ide-sources`, or `null` when Gradle is omitted. |
| `packages` | The launcher and optional source helper, ready for `mkShell`. |
| `shellEnv` | `GRADLE_HOME` when Gradle is supplied; otherwise empty. |

Merge `shellEnv` into `mkShell` to support Eclipse/Buildship launched directly
from that shell, including projects without a Gradle wrapper. The JDTLS launcher
uses the same Gradle home. For local tool development, you can temporarily use a
local flake input or a Nix input override instead of the GitHub input.

## Use with Helix

Activate the project's devShell before starting Helix and keep the Java server
command as `jdtls`. Reload the environment and restart Helix after updating the
input. No project-local `languages.toml` is needed for the source capability.

Workspaces live under `${XDG_CACHE_HOME:-$HOME/.cache}/jdtls/`, keyed by the SHA-256
of the Git checkout root. A non-Git project uses its physical working directory.
Separate checkouts with the same name get distinct caches; subdirectories of one
checkout share its workspace. Set `JDTLS_WORKSPACE` to choose a workspace explicitly,
for example when running separate language-server sessions in one checkout.

The launcher preserves arguments and exposes the real server to the proxy without
command-name recursion. Its private read-only source cache uses the operating
system temporary directory. JDTLS project import and source attachment must still
succeed; this bridge does not replace the language server's build-system support.

## Resolve Gradle dependency sources

With Gradle supplied, run from the build root:

```sh
java-ide-sources
```

The helper covers the root Java project and Java subprojects, supporting both
single-module and multi-project builds using Gradle's Java plugins. It uses the
Eclipse classpath model without generating IDE project files. Non-Java aggregator
projects are skipped. The helper does not provide Android-specific import support.

After an intentional dependency update, generate missing verification entries:

```sh
java-ide-sources --write-verification-metadata sha256
```

Review generated checksums against canonical artifacts before accepting them.
The helper respects project dependency verification and does not change versions
or locks. Source checksums and fixed dependency-cache maintenance belong to the
project. Eclipse/Buildship and JDTLS can both use those verification entries.

## Android Gradle projects

Use the separate Android helper with the project's existing build tools:

```nix
javaEditor = java-dev-tools.lib.mkAndroidEditor {
  inherit pkgs jdk gradle;
  # Optional: Nix-patched AAPT2 matching the project's SDK/build-tools selection.
  aapt2 = "${sdkRoot}/build-tools/36.0.0/aapt2";
};
```

Merge `javaEditor.shellEnv` and install `javaEditor.packages` as in the Java
example. No project `.helix/languages.toml`, copied Gradle init script or launcher
override is needed. The same repository owns both helpers; Android is not enabled
by `mkJavaEditor` or the standalone default package.

`mkAndroidEditor` defaults to a private `pkgs.jdk21_headless` server runtime,
keeping the supplied build JDK in `JAVA_HOME`. It exports the supplied Gradle
installation through `GRADLE_HOME`. The proxy supplies JDTLS startup settings
before the first import and preserves the adapter on configuration updates.
Other client preferences override defaults; client Gradle arguments are retained
alongside the required init script, without duplicating it.

The opt-in adapter provides AGP 9 application/library source folders, dependencies,
SDK jars, generated R classes and separate IDE output under `build/eclipse`.
It selects one unflavored build type (`variant = "debug"` by default), with JVM
tests (`unitTests = true`). Product flavors, instrumented tests and Kotlin source
editing are outside this model. An unsupported/flavored selection fails explicitly
instead of combining incompatible source roots. A custom Gradle build directory
is respected. Normal Gradle builds do not load this adapter.

The optional `aapt2` argument selects a patched executable for Buildship; it does
not depend on the CLI's `GRADLE_OPTS`. Omit it when the project's normal Maven
AAPT2 works. Dependency source downloads default to false; use
`downloadSources = true` only when the project's verification policy accepts those
archives. Binary dependencies retain their normal resolution and strict
verification. The generic `java-ide-sources` command is omitted in Android mode.
`androidInitScript` is exposed for direct model diagnostics.

The shared launcher also accepts `initializationSettings = { ...; }` when called
through `mkJavaEditor`. Empty settings (the default) leave ordinary Java startup
and client configuration unchanged apart from the existing source capability.
Both nested and dotted JDTLS preference keys are accepted by the merge.

## Maintain this flake

Standalone packages use the committed nixpkgs lock and default to JDK 25 and
Gradle 9. The library API uses the caller's packages and toolchains. Packages are
exposed on x86_64 Linux, aarch64 Linux and aarch64 Darwin: `jdtls` (default),
`jdtls-proxy`, and `java-ide-sources`.

The human-owned qualification command is:

```sh
nix flake check
```

It builds/tests the proxy on the current system. Protocol tests cover virtual-source
translation, overlapping request IDs, text escaping, reverse URI mapping, request
preservation, read-only cache replacement, automatic capability negotiation and
shutdown, opt-in startup settings, client preference preservation and later
settings updates. An offline Gradle check covers root and multi-project Java
source helpers and verifies that the Android adapter leaves Java models unchanged.
Actual editor navigation and Android Gradle project import are separate
integration checks. Automated agents must not run Nix.

The upstream proxy is [quantonganh/jdtls-wrapper](https://github.com/quantonganh/jdtls-wrapper)
(MIT), pinned to commit `51df1676f9d0d1116a5a14f2d3741a9a425202e4`. This flake owns
the package, correctness patch and tests. The patch also enables the source
capability implemented by the bridge. Nix supplies the Go toolchain; the proxy
uses only the Go standard library.
