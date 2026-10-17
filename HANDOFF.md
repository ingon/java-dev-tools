# Extraction handoff

This directory is the complete extracted `java-dev-tools` repository. It was
moved into `totipo-java/java-dev-tools/` so it is inside the persistent workspace
mount. Move the entire directory, including its `.git` directory and `.gitignore`,
to your intended host location before committing and pushing it.

The repository was initialized on `main`, but has no commits and no staged
contents. No remote has been configured. The parent Totipo repository ignores
this transport directory to prevent accidentally recording a nested Git repository.

## Included

- A standalone flake and an inherited nixpkgs lock (same revision already used
  by Totipo; no toolchain update).
- `lib.mkJavaEditor { pkgs; jdk; gradle ? null; }`, returning `jdtls`, the proxy,
  optional `java-ide-sources`, and a `packages` list for the caller's devShell.
- The pinned upstream proxy package, its correctness patch and Go protocol tests.
- Automatic source-capability negotiation; consuming projects need no Helix
  language configuration file.
- Workspace selection keyed by checkout-root hash, with `JDTLS_WORKSPACE` override.
- The shared Gradle source-resolution init script and command.
- README usage and maintenance instructions.

The standalone flake owns its proxy check. Totipo no longer contains the Go
package definition, proxy patch/tests, project-local Helix setting or source-helper
implementation. Its dependency checksums, Eclipse output layout and ignore rules
remain project-owned. `totipo-desktop` is not present in this workspace and has
not been modified; use the same README integration there.

## Validation status

Before transport, the extracted patched proxy and all five Go test functions
passed using the existing temporary Go 1.25.14 validation toolchain, with module
downloads disabled. The new test covers automatic capability negotiation and
preservation of existing initialization fields/settings. The prior proxy's
navigation, escaping, overlapping-ID, request preservation and source-cache tests
also passed. No Nix command was run by the agent.

The new generic launcher and the flake/library composition have not yet been
qualified by Nix. Workspace-isolation and extracted command checks remain to be
finished after reconnection. This is a reviewable extraction, not a claim of a
completed, qualified migration.

## Resume after publishing

Give the agent the remote URL and access to the new repository. In `totipo-java`,
replace the provisional `path:./java-dev-tools` input with the remote flake URL,
keeping `inputs.nixpkgs.follows = "nixpkgs"`. Its current `flake.lock` has not yet
been updated for this new input. Review that lock update, finish focused checks,
and run the appropriate human Nix gates before treating the extraction as qualified.
Remove the temporary parent ignore rule after moving this directory out.

The earlier three reported successful Totipo Nix gates predate the extraction;
included-input changes mean they do not qualify the current checkout. The Java
dependency cache and source checksums have not changed during extraction.
