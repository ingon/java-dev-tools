# Java development tools for Nix

Reusable JDTLS source navigation for Helix and a Gradle dependency-source helper.
This flake owns the pinned proxy package, correctness patch and protocol tests;
application/library repositories only consume its tools.

The `jdtls` command runs the Nix JDTLS server through a source-document bridge.
The bridge advertises its supported client capability during initialization, so
Helix's built-in Java configuration works without project-specific language
settings. JDK and dependency source archives must still be present and accepted
by the project's Gradle dependency verification policy.

## Consume from a project

While developing sibling checkouts, add the deliberately local path input:

```nix
inputs.java-dev-tools = {
  url = "path:../java-dev-tools";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

Add `java-dev-tools` to the consuming flake's `outputs` arguments. Given the
project's existing `pkgs`, `jdk` and `gradle`, compose its development shell:

```nix
let
  javaEditor = java-dev-tools.lib.mkJavaEditor { inherit pkgs jdk gradle; };
in {
  devShells.default = pkgs.mkShell {
    JAVA_HOME = "${jdk}/lib/openjdk";
    packages = [ jdk gradle ] ++ javaEditor.packages;
  };
}
```

The helper uses the caller's packages and toolchains, including overriding JDTLS
to run on the caller's JDK. It does not import a second nixpkgs for that use.
`gradle` is optional; omit it for a Maven/non-Gradle project to get only `jdtls`.
Returned attributes are `jdtls`, `proxy`, `sources` (`null` without Gradle), and
`packages` (the launcher plus the optional source command). Project-specific
build/publication/check definitions stay in the consuming repository.

This is the same integration for `totipo-java` and `totipo-desktop`; no Go source,
patch, tests, Gradle helper file or Helix configuration needs copying. After a
remote repository is created and this work is published, replace the local URL
with that repository's `github:OWNER/REPO` URL and update only this input's lock.
A remote input avoids the sibling-checkout requirement and is pinned by the
consumer's lockfile. No remote repository has been created by this extraction.
Local path inputs intentionally include the local source tree; keep build and
cache artifacts out of this repository.

## Commands

Activate the project devShell before starting Helix. Keep its server command as
`jdtls`. Reload the environment and restart Helix after updating this input.

The launcher selects a workspace under `${XDG_CACHE_HOME:-$HOME/.cache}/jdtls/`,
keyed by the SHA-256 of the Git checkout root (or physical current directory for
non-Git projects). Separate checkouts with the same basename use distinct caches;
starting in a subdirectory of the same checkout reuses that checkout's workspace.
Set `JDTLS_WORKSPACE` to choose an explicit workspace. The launcher preserves
caller arguments and exposes the real server to the proxy without name recursion.
Its private read-only source-file cache uses the operating system temporary path.

With Gradle supplied, verify/download dependency sources from a project root:

```sh
java-ide-sources
```

For an intentional dependency update, generate missing checksum entries with:

```sh
java-ide-sources --write-verification-metadata sha256
```

Review every generated checksum against canonical artifacts. The command retains
strict verification by default and does not change dependency versions or locks.
Source checksums and any fixed Gradle dependency-cache maintenance remain the
consuming project's responsibility. Eclipse/Buildship and JDTLS can both consume
those checked-in checksums, independently of this editor bridge.

## Package maintenance

Standalone outputs use the committed nixpkgs lock, inherited unchanged from the
original consuming project; consuming projects can follow their own nixpkgs.
Packages are exposed on x86_64 Linux, aarch64 Linux and aarch64 Darwin:
`jdtls` (default), `jdtls-proxy`, and `java-ide-sources`. Standalone defaults use
JDK 25 and Gradle 9; the library API accepts the consumer's chosen supported tools.

The human-owned qualification command is:

```sh
nix flake check
```

It builds/tests the proxy on the current system. Its tests cover virtual-source
translation, overlapping request IDs, text escaping, reverse URI mapping,
request-field preservation, read-only cache replacement, automatic capability
negotiation and shutdown. The proxy uses only the Go standard library. Tests
operate against a protocol fixture; actual JDTLS/editor navigation is a separate
integration check after environment reload. Automated agents must not run Nix.

Upstream is [quantonganh/jdtls-wrapper](https://github.com/quantonganh/jdtls-wrapper)
(MIT), pinned to commit `51df1676f9d0d1116a5a14f2d3741a9a425202e4`. The local
patch fixes request ID allocation, escaping, message preservation, synchronization,
pipe termination and source-cache handling, and enables the capability the bridge
implements. This repo packages/patches that upstream project rather than vendoring
or forking its Go module. Nix supplies the Go toolchain during package builds.
