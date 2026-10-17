# Java editor toolkit implementation and validation

This report owns reusable editor implementation/history. Dependency checksums,
project output layout and consumer qualification belong to each consuming project.
The originating consumer retains one integration report.

Origin development started on `totipo-java` main at
`952c7c3a2ec2115063a1a602254d8f7ed8d77e44`, preserving existing user changes.
Initial toolkit publication was `e8c2e6dc0cc99066e74a99bc3109a2854c30aa60`.
Shared environment work started from that clean revision and was published by
the user as `0685f64189151a77d0c399441c0dffb311541ec2`. Generic integration/helper
work started from clean main at the latter revision. Classifications:
`DEPENDENCY_BUILD` for editor packaging/helper behavior; `DOCS_JAVADOC` for
Markdown documentation. This report consolidation changes only Markdown.

## Implementation ownership

`lib.mkJavaEditor { pkgs; jdk; gradle ? null; }` exports `jdtls`, `proxy`,
`sources`, `packages` and `shellEnv`, using caller-owned packages/toolchains.
Optional Gradle adds `java-ide-sources` and `shellEnv.GRADLE_HOME`; the launcher
uses the same value. Without Gradle, `sources` is null and `shellEnv` is empty.

The launcher exports the chosen JDK, hashes the Git checkout root for isolated
workspaces (physical working directory outside Git), respects `JDTLS_WORKSPACE`,
preserves arguments and puts the real server first on PATH to avoid recursion.
The proxy automatically advertises `classFileContentsSupport`, preserving other
initialize options. No project Helix file is required. Virtual sources become
ordinary read-only files for editor navigation.

The proxy pins `quantonganh/jdtls-wrapper` commit
`51df1676f9d0d1116a5a14f2d3741a9a425202e4`. Archive SHA-256:
`51caa30edf67b41ba59c32700d77dcef6f0c47dcb9467c5f4959236eb79d5db0`
(SRI `sha256-UcqjDt9ntBulnDJwDXfc728MR9y5RnxfSVkjbredXbA=`).
Archive `main.go`/`go.mod` matched a separate clone of the commit. There are no
external Go modules (`vendorHash = null`); Nix provides the build toolchain.
This fixed source is independent of consumer Gradle caches.

The patch reserves negative internal IDs, preserves source escapes after one
JSON decode and all request fields during URI reverse mapping, synchronizes
shared URI state/pipe writes, handles EOF and uses a private per-process source
cache with atomic read-only replacement. The zero-context patch reapplies with
`git apply --unidiff-zero` and produced byte-identical tested source. Package
`doCheck` runs the copied tests; `checks.<system>.proxy` selects that package.

The opt-in Gradle helper resolves Eclipse attachments without emitting project
files and disables configuration-cache reuse. The local generic follow-up fixes
silent omission of root Java projects: visit all projects with the standard Java
plugin, including root, skipping non-Java aggregators. This does not claim
Android or arbitrary build-system support.

README describes JDTLS-supported projects and the public API.
[INTEGRATION.md](../INTEGRATION.md) gives humans/agents flake composition,
Gradle/non-Gradle variants, optional Kotlin/Groovy Eclipse layout, ignore rules,
source verification and human lock/qualification steps. Project output layout
belongs to the consumer. Obsolete transport `HANDOFF.md` is removed; test fixture
names/environment variables are generic.

## Validation and invocation accounting

Agents ran no Nix. Focused Go work used an auxiliary official Go 1.25.14 archive,
verified SHA-256
`a21ae5633a269bcd7e90cf767e48225633795e99d831742cbf3397064fee7712`.
It does not change the flake toolchain. Tests/builds used `GOTOOLCHAIN=local`,
`GOPROXY=off` and temporary external caches.

Initial proxy adoption: six Go test invocations:

1. Wrong-directory fixture copy left no tests: not validation evidence.
2. Early patched integration pass.
3. Unmodified upstream failed on editor/internal ID collision.
4. Cached patched pass: no fresh execution.
5. Fresh final patched four-function suite passed after cache coverage.
6. Fresh exact-pinned-archive patched suite passed.

Two auxiliary Go builds passed. Two patch-preparation assertions failed before
generation, changing no repository inputs and providing no passing evidence.
The four functions were two child-process fixtures and substantive navigation/
cache tests: overlapping IDs, escaped text, complete reverse-mapped requests,
URI mapping, shutdown and read-only replacement. A temporary launcher harness
checked Bash syntax, source pin/hash, initial capability TOML, real-server lookup
without recursion, workspace paths with spaces and exact forwarding. Evidence:
`/tmp/totipo-jdtls-proxy-audit/`. Automatic capability injection superseded TOML.

Extraction: one fresh Go test invocation passed all five functions, adding
capability negotiation coverage. One auxiliary Go build supported ten published
launcher process executions covering distinct same-basename checkouts, shared
root/subdirectory workspace, non-Git fallback, explicit override, spaces/args,
with/without Gradle, JDK/Gradle exports and nonrecursive lookup. The actual shell
template was rendered with controlled paths; this was not Nix evaluation.
One published-helper Gradle invocation on the original multi-project consumer
executed one task and attached 12/11 sources under strict verification. Evidence:
`/tmp/totipo-java-editor-published-audit/`.

Nine transferred source/document/lock files matched initial published revision
byte-for-byte. Its nixpkgs pin matched the consumer; no toolchain update occurred.
Temporary nested transport was a workspace-mount workaround. The user moved and
published the repository; no agent created a remote, committed or pushed.

The `shellEnv` follow-up changed only library/README and was reviewed for both
optional-Gradle branches. No new Go/build/Java tests were added. The user confirmed
consumer Helix navigation into project, JDK and Bouncy Castle sources and passed
its final Nix gate at `0685f64`. Consumer gate counts remain in its integration
report; they are not standalone toolkit Nix claims.

The generic documentation/root-project follow-up used five Gradle invocations,
seven executed tasks: root fixture baseline (task succeeded with no sources),
corrected root fixture (one strict source attachment), real multi-project run
(12/11 attachments), and exact Kotlin/Groovy guide recipes. Both recipes checked
root/child default/per-source output paths, including a customized build dir.
The two recipe runs were offline without external dependencies, not full offline
qualification. No JUnit XML was overwritten. One fresh Go invocation passed all
five functions after fixture renaming, with verified source/compiler archives;
no proxy implementation or toolchain pin changed. Evidence:
`/tmp/java-dev-tools-generic-audit/`.

Across these three phases: eight Go test invocations, including the failed,
no-test and cached distinctions above; three auxiliary Go builds. Broad Java
baselines, focused Java unit/API and standalone release-tool tests, full Java
gates, publication/consumer passes, separate host full offline suites and release
reproducibility runs: all zero. Original consumer source/layout Java gates are
accounted separately in that repository.

## Qualification and final Git state

Generic helper/docs changes are unpublished. The toolkit's human-owned
`nix flake check` is pending before publishing implementation changes; no
standalone toolkit Nix pass is claimed. The consumer's immutable old pin is
unchanged and remains qualified. Report consolidation needs no new Java/Nix
runs: package/check inputs reference implementation paths, not `review/`.

Final `git status --short` and `git diff --name-only` contain deletion of
`HANDOFF.md`, modifications to README, `gradle/ide-sources.gradle`,
`tests/jdtls-wrapper_test.go`, and intent-added `INTEGRATION.md` plus this report.
`git diff --check` and `git diff --cached --quiet` pass;
`git diff --cached --name-only` is empty. No generated output/download/cache or
temporary evidence is Git-known. No actual contents were staged or committed.
