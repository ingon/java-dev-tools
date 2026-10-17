# Shared Android editor integration

Starting state: clean main at b5fe56eaebc570a50dbb9803e7d7a8cf179731d3.
Classification: BUILD_DEPENDENCY and TEST_OR_QUALIFICATION_TOOLING, with docs.
User authorized integrating the working Totipo Android editor support here,
preserving ordinary Java integration, and will perform publication personally.
No new repository, commit, push or actual staged contents. Agent Nix runs: zero.

## Final API and isolation

`lib.mkJavaEditor` retains its existing required/optional arguments and defaults.
`serverJdk` optionally separates server runtime from build JDK, defaulting to jdk.
`initializationSettings` defaults to empty: ordinary Java startup/settings do not
get Android preferences. The existing Gradle source helper is byte-identical.
Existing standalone package names/defaults and all dependency pins are unchanged.

`lib.mkAndroidEditor` is a separate opt-in composition using the same launcher
and proxy. It retains build jdk/JAVA_HOME and supplied Gradle/GRADLE_HOME, defaults
serverJdk to caller pkgs.jdk21_headless, and owns the Android init script/startup
settings. Parameters: variant (debug), unitTests (true), aapt2 (null), and
 downloadSources (false). It exposes androidInitScript and returns only jdtls in
packages; sources is null because the generic Java source helper is not applicable.

The Android model covers one unflavored AGP 9 application/library build type and
its JVM tests, SDK jars, dependencies and generated R classes. Instrumented tests,
product flavors and Kotlin editing are not modeled. A flavored/invalid selection
fails explicitly. Tests below cover this Java-source app with JAR dependencies
and an isolated library fixture, not every Android dependency/module layout.
IDE output follows each module's build directory. Optional patched AAPT2 is set
before Android plugin application for Tooling API/resource synchronization.
Normal project builds and mkJavaEditor do not load the adapter.

A packaged standard-library Go helper merges opt-in JDTLS preferences before the
first import, and on later workspace/didChangeConfiguration. It handles nested
and dotted preference keys, preserves unrelated request/options/capability fields,
lets client scalar preferences override defaults and retains both required init
arguments and client Gradle arguments without duplication. Invalid environment
JSON fails startup. Existing source-capability/cache/URI behavior is retained.
No preferences are injected when defaults are empty.

README and INTEGRATION describe the distinct helpers, supported model, source
checksum policy and human update/validation steps. Android consumers need no
copied Gradle script, Helix language settings or launcher text override.

## Validation and limitations

- Plain root Java and multi-project Java regression fixtures: PASS twice. Each
  run performs four offline Gradle invocations, exercising the unchanged source
  helper with and without the Android adapter. Source roots and complete generated
  classpath XML remain identical; both projects resolve zero source attachments.
  These fixtures are now the java-models Nix check, using exact source inputs.
- Extracted Android app model/resource generation: PASS on Totipo AGP 9.4.0,
  Gradle 9.8.0 and JDK 17. Exact four intended source roots and platform/Totipo/
  JUnit/generated R libraries verified. Final focused classpath rerun PASS.
- Isolated AGP 9.4 Android library model/resource generation: PASS; additional
  offline run with unitTests=false and a custom build directory PASS. Asserted
  only main/debug roots, no JUnit, and IDE output under custom-output/eclipse.
- Standalone Tooling API EclipseProject query of the extracted adapter: PASS,
  asserting the same four roots and real Android/Totipo/JUnit/R classpath files.
- New Go unit/protocol regressions cover default/invalid environment settings,
  nested/dotted options, client precedence, required argument retention, later
  configuration updates, no-duplicate adapter, unchanged ordinary Java settings,
  source navigation, fields and unknown capability preservation.
  Go compiler is unavailable in the agent environment. Human reports the
  final `nix flake check` passed, covering these packaged tests and Java-model
  regressions. No independent agent Go execution or per-test log is claimed.
- Patch regenerated against the pinned upstream wrapper and checked locally for
  consistent application/source. git diff --check PASS. Gradle source-helper and
  flake.lock byte-identity confirmed; no verification/lock/cache edits here.

One initial extracted-adapter offline invocation failed on an uncached AGP marker
in the new agent environment. A narrow online retry passed without source or
verification changes. One Tooling API harness launch had a mistyped classpath
and failed compilation before running; corrected invocation PASS. These were
focused diagnostics, not repeated full gates. Human Nix evaluation/package/check qualification passed. Actual editor
navigation through the new shared launcher remains pending publication/migration.
The previously working local Totipo integration is preserved until publication.

## Freeze and human handoff

Production/source inputs stable: shared launcher/helper/adapter and proxy changes
reviewed. Tests stable: Java fixtures passed; Go regressions prepared for the
human build. Build/package configuration stable: new check and packaged helper,
existing pins unchanged. Qualification tooling stable: focused fixtures/report
complete. New files receive intent-to-add membership only before human Nix.
No staged-content diff; no ignored diagnostic/build/cache state enters Git.

Frozen file SHA-256 (report excluded from effective checks):

```text
36090bfa2235c6c8f8da3e851950602345b45ae52ec8bc2effc1369d76b77e24 INTEGRATION.md
fb3f6426f121d3027c0014a7f6918327bf37d807027e41836d3b88d10460ad72 README.md
a375284dffbd8beef5a88be9d82cb872190654bdfb72c86de6b7890b89e568c1 flake.nix
37377514dda3d31084352432e6307929e3f710f640c9fcebea3e3b14fedf4b98 lib/mk-java-editor.nix
fdf1cdeb87e991d622807bf4210767fdd9b3e2d7040e9811cf0286f5fc9cfa88 lib/mk-android-editor.nix
9de0704b2b7783770d91bca858537050a539f3cd35cb5dadfcb3aa970b10131a gradle/android-ide.init.gradle
92fd34aaff175ad1f358f5cbb781aa44b396e4f4071056f91bf11e2ab0d6c441 pkgs/jdtls-proxy.nix
03b1ea4147ad99a2f898008fa8471a8ac3e41f7a62de1c89a50d0fbd5b3cc277 pkgs/jdtls-wrapper.patch
698d6457da76ba8ea3261492e4b148d8f75572e95f9c7d4ed1529796aee5723e pkgs/initialization-settings.go
adf1540c95fc22ae91be95ba538382d1e7e748893e99a6fdc1c3d956259e6ca8 tests/jdtls-wrapper_test.go
fd6ce013561d0d19031550c254d332108a77ab12f796ab2ec62f017a8d78f6e8 tests/initialization-settings_test.go
6cd7e1defc270270fb9d2b73f9ef69db7b5ac269f2bf19ad15028dd2874e18fc tests/gradle-models.py
```

Human `nix flake check`: PASS, user-reported. Supplied output contained only
dirty-tree and incompatible-system omission warnings, followed by explicit
confirmation that all checks passed. No duration/task/cache breakdown supplied.
All recorded source hashes remain unchanged afterward; only this excluded
report is updated, so the result is not invalidated. Agent Nix executions: zero.
Next: user commits/pushes the qualified tooling changes. Then Android can update only its
java-dev-tools input, switch to mkAndroidEditor, remove local .helix files and
launcher workaround, recheck editor navigation and run its final human Nix gate.
A concrete consumer-simplification draft is saved in ignored Android diagnostics;
Android's currently working flake/local integration is not changed prematurely.

## Counts and Git state

- Direct focused Gradle CLI during this extraction: 5 (one offline cache failure,
  app/resource success, library success, library custom-directory/no-tests success,
  final app model success).
- Plain Java regression commands: 2 script runs, 8 nested offline Gradle calls PASS.
- Standalone Tooling API queries: 1 PASS; separate initial harness compile typo.
- Full Gradle/clean/APK-verifier/device runs: 0; editor-only changes.
- Human/agent toolkit Nix checks: 1 PASS reported / 0.
- Toolkit lock/dependency/verification changes: 0.

All 13 changed/new paths remain unstaged/uncommitted on starting main. This report
is an excluded evidence file: checks reference the specific helper/template/test
sources, not the whole repository, so report-only result updates do not invalidate
the toolkit gate. No published revision or final Android migration is claimed yet.
