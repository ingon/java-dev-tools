# Integrate a Java project

Use this guide for a human or agent integrating an existing project. The toolkit
supports Java projects that JDTLS can import; it is not tied to an application or
library. The Gradle helper supports root Java projects and Java subprojects using
Gradle's Java plugins. Maven and other supported layouts can use the launcher
without that helper. Android projects use the separate opt-in helper below.

Read the target repository's `AGENTS.md` and build instructions first. Preserve
its JDK, Gradle version, compilation target, locks and verification policy. This
integration does not require adding runtime dependencies or copying any Go code,
proxy patches, tests or Gradle scripts into the target project.

Automated agents should prepare and review the changes below, but must not run
Nix commands. The Nix commands in this guide are human steps. Do not commit,
push or stage file contents unless the target repository/user authorizes it.

## 1. Inspect the existing layout

Identify the target flake's per-system `pkgs`, JDK and optional Gradle definitions,
its existing devShell and package list, and whether it uses plain flake outputs,
flake-utils or flake-parts. Keep that structure.

For Gradle, identify the root build script, Java modules, existing Eclipse model
configuration and ignored output directories. A single-module root project needs
the same output configuration as a Java subproject. For Maven, skip step 3 and
the Gradle source-helper commands.

## 2. Update `flake.nix`

Add the input, following the target project's nixpkgs:

```nix
java-dev-tools = {
  url = "github:ingon/java-dev-tools";
  inputs.nixpkgs.follows = "nixpkgs";
};
```

Add `java-dev-tools` to the flake's `outputs` arguments, or access it through an
existing `inputs` binding. In the existing per-system scope where `pkgs`, `jdk`
and `gradle` are defined, add:

```nix
javaEditor = java-dev-tools.lib.mkJavaEditor { inherit pkgs jdk gradle; };
```

Merge the exported environment into the existing shell and append its packages.
The following shows a minimal per-system shell; retain the target's other
packages, environment attributes, hooks and sandbox settings:

```nix
devShells.default = pkgs.mkShell (javaEditor.shellEnv // {
  JAVA_HOME = "${jdk}/lib/openjdk";
  packages = [ jdk gradle ] ++ javaEditor.packages;
});
```

If the flake defines shells directly at the top level, the key will instead be
`devShells.${system}.default`. In flake-parts, use the existing `perSystem` shell.
Do not introduce a second system loop or replace the flake's build/check outputs.

Remove a duplicate manually constructed `GRADLE_HOME` assignment: `shellEnv` owns
that setting and also supports Eclipse/Buildship started from the shell. Keep
any project-specific variables, such as an explicit Gradle executable used by CI.

For Maven/non-Gradle use, omit the `gradle` argument and package from these examples
and install the project's build tool separately. `shellEnv` is then empty and
`javaEditor.packages` contains only the JDTLS launcher. The supplied JDK must be
supported by the selected JDTLS version; do not change the build's bytecode target
merely to select a language-server runtime.

The project needs no new `.helix/languages.toml` for either toolkit helper. Keep unrelated
editor settings. If there is a custom Java server command, ensure it invokes
`jdtls` from the devShell rather than bypassing the bridge with an absolute server
path or an older proxy. Do not hardcode a shared `-data` workspace; the launcher
isolates checkouts automatically and accepts `JDTLS_WORKSPACE` when needed.

### Android instead of plain Java

For an AGP 9 application/library with an unflavored build type, use:

```nix
javaEditor = java-dev-tools.lib.mkAndroidEditor {
  inherit pkgs jdk gradle;
  aapt2 = "${sdkRoot}/build-tools/36.0.0/aapt2"; # only if a patched AAPT2 is needed
};
```

Keep the project's JDK/Gradle/SDK selection and use its actual build-tools version
in `aapt2`. The helper defaults the server runtime to JDK 21; `serverJdk` can
override that without changing the build JDK. `variant` defaults to `"debug"`,
`unitTests` to true and `downloadSources` to false. Merge `shellEnv` and install
`packages` exactly as in step 2. Skip step 3: this helper already isolates IDE
output in the selected module's build directory. Skip the standard Java source
helper in step 5; it is not installed in Android mode. Preserve strict verification
and review source-archive checksums before enabling their downloads.

Do not copy the adapter into the consuming project or add Android startup
preferences to Helix: the helper and proxy own both. When migrating a local
integration, remove those now-redundant files/launcher patches only after switching
to the new toolkit revision. Existing Java consumers keep `mkJavaEditor`; their
defaults, source helper and import behavior remain unchanged.

The Android model covers one unflavored application/library build type plus its
JVM tests. It does not cover instrumented tests, product flavors or Kotlin editing.
Validate actual app import and app/SDK/dependency navigation after environment
reload; server discovery alone is insufficient. A fresh temporary JDTLS workspace
can distinguish stale import state from adapter failure without deleting caches.

## 3. Optionally update `build.gradle.kts` output paths

This step keeps Eclipse/JDTLS compiler output in an ignored build directory,
separate from Gradle's compiled classes. It is a project build-model choice,
not required for the source-navigation bridge itself. Skip it if the project
already has a suitable separate IDE output location.

After any existing build-directory customization in the root Kotlin build script,
configure the model only when the Eclipse plugin
is applied by Buildship/JDTLS. Do not apply the plugin to every normal build just
to make this callback run:

```kotlin
allprojects {
    pluginManager.withPlugin("eclipse") {
        extensions.configure<org.gradle.plugins.ide.eclipse.model.EclipseModel> {
            classpath.defaultOutputDir = layout.buildDirectory.dir("eclipse/default").get().asFile
            classpath.baseSourceOutputDir.set(layout.buildDirectory.dir("eclipse"))
        }
    }
}
```

`allprojects` covers both a root Java project and Java subprojects. Do not use only
`subprojects` for a single-module build. If an existing `allprojects`/`subprojects`
block or convention plugin already configures the Eclipse model, adapt that code
instead of adding duplicate or conflicting configuration. The
[`baseSourceOutputDir`](https://docs.gradle.org/current/javadoc/org/gradle/plugins/ide/eclipse/model/EclipseClasspath.html#getBaseSourceOutputDir())
requires Gradle 8.1 or later; preserve a compatible
existing approach on older builds rather than upgrading Gradle just for this.

This results in module-local `build/eclipse/main`, `test`, `testFixtures` when
present, and `build/eclipse/default`, relative to each project's build directory.
A project that customizes `layout.buildDirectory` keeps that customization.
Gradle's own class directories, test configuration and publication stay unchanged.

For a Groovy build script, the equivalent callback is:

```groovy
allprojects {
    pluginManager.withPlugin('eclipse') {
        eclipse.classpath.defaultOutputDir = layout.buildDirectory.dir('eclipse/default').get().asFile
        eclipse.classpath.baseSourceOutputDir.set(layout.buildDirectory.dir('eclipse'))
    }
}
```

## 4. Update `.gitignore`

Add only missing generated-state/output patterns appropriate to the project:

```gitignore
.direnv/
**/build/
.project
.classpath
.settings/
```

Preserve intentional checked-in IDE configuration. If the project uses a custom
build directory, keep its corresponding ignore pattern. If output previously went
to module `bin/` folders, refresh the Gradle import before removing old generated
output. Any temporary ignores should name those module paths specifically;
a global `bin/` rule can hide executable scripts.

Ignore generated state; do not commit `.project`, `.classpath`, `.settings/`,
compiler output, language-server workspaces or downloaded caches as part of this
integration. Do not alter the project's Nix source filter to include them.

## 5. Handle source archives and verification

The JDK should provide its source archive. Dependency source JARs must be available
to the build model as well; a checksum for a binary JAR does not cover its sources.

For Gradle, from the build root in the activated devShell:

```sh
java-ide-sources
```

If verification rejects previously unrecorded source JARs, intentionally generate
their entries with:

```sh
java-ide-sources --write-verification-metadata sha256
```

Review the metadata diff and validate the new checksums against canonical artifacts
before accepting them. Keep existing checksums and dependency versions unchanged.
Do not add a blanket verification exemption or disable verification just to make
an IDE import succeed. Follow the target's additional fixed-cache maintenance
policy if its dependency inputs changed. For Maven, use its source-attachment and
dependency policy rather than the Gradle command.

## 6. Lock, validate and hand off

For a new flake input, the human adds its lock without updating existing pins:

```sh
nix flake lock
```

For a deliberate later toolkit update, the human targets that input:

```sh
nix flake update java-dev-tools
```

Review the new revision and ensure its nixpkgs follows the existing root input.
Keep all unrelated toolchain pins unchanged. Run the target's required checks for
the files changed, respecting its qualification policy. For this toolkit's own
implementation changes, its human-owned package/protocol gate is `nix flake check`
from the toolkit repository; a consumer's checks need not own the proxy's tests.

For an editor smoke check, the human reloads direnv or enters `nix develop`, then
restarts Helix from that environment. Use `hx --health java` to verify discovery,
then test navigation within the project and into a JDK/dependency class. Health
output alone does not prove project import or source navigation succeeded.
Refresh the Gradle import in Eclipse/JDTLS to adopt output-layout changes.

A complete handoff identifies changes to the flake, optional Gradle model,
ignore rules and any source-verification entries, plus the locked toolkit revision,
checks performed, human steps still pending and unchanged project toolchains.
Do not claim an unperformed Nix run or actual editor smoke test as passed.
