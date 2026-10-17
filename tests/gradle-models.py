"""Offline Java regression fixtures for the shared helper and Android no-op path."""

import argparse
from pathlib import Path
import subprocess
import tempfile
import xml.etree.ElementTree as ET


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--helper", type=Path, required=True)
    parser.add_argument("--adapter", type=Path, required=True)
    args = parser.parse_args()
    helper = args.helper.resolve()
    adapter = args.adapter.resolve()
    for multi in (False, True):
        with tempfile.TemporaryDirectory(prefix="java-editor-regression-") as temporary:
            root = Path(temporary)
            (root / "settings.gradle").write_text(
                "rootProject.name = 'fixture'\n" + ("include 'child'\n" if multi else "")
            )
            modules = [root, root / "child"] if multi else [root]
            for module in modules:
                module.mkdir(exist_ok=True)
                (module / "build.gradle").write_text("plugins { id 'java' }\n")
                for source in ("main", "test"):
                    (module / f"src/{source}/java").mkdir(parents=True)
            (root / "gradle.properties").write_text(
                "org.gradle.java.installations.fromEnv=JAVA_HOME\n"
                "org.gradle.java.installations.auto-download=false\n"
                "org.gradle.java.installations.auto-detect=false\n"
            )
            snapshots = []
            for extra in ([], ["--init-script", str(adapter)]):
                command = [
                    "gradle", "--offline", "--no-daemon", "--no-configuration-cache",
                    "--init-script", str(helper), *extra, "ideSources", "eclipseClasspath",
                ]
                result = subprocess.run(command, cwd=root, text=True, capture_output=True)
                if result.returncode:
                    raise RuntimeError(result.stdout + result.stderr)
                assert result.stdout.count("resolved 0 dependency source attachments") == len(modules)
                snapshot = []
                for module in modules:
                    entries = ET.parse(module / ".classpath").getroot()
                    sources = {e.get("path") for e in entries if e.get("kind") == "src"}
                    assert sources == {"src/main/java", "src/test/java"}, sources
                    assert not any("android" in (e.get("path") or "").lower() for e in entries)
                    snapshot.append(ET.tostring(entries))
                snapshots.append(snapshot)
            assert snapshots[0] == snapshots[1], "Android adapter changed a Java project model"
    print("PASS: root Java and multi-project Java source helpers; Android adapter leaves both models unchanged")


if __name__ == "__main__":
    main()
