{ lib, buildGoModule, fetchurl, runtimeShell }:
buildGoModule {
  pname = "jdtls-wrapper";
  version = "0-unstable-2025-11-24";
  src = fetchurl {
    name = "jdtls-wrapper-51df1676.tar.gz";
    url = "https://codeload.github.com/quantonganh/jdtls-wrapper/tar.gz/51df1676f9d0d1116a5a14f2d3741a9a425202e4";
    hash = "sha256-UcqjDt9ntBulnDJwDXfc728MR9y5RnxfSVkjbredXbA=";
  };
  # Only the Go standard library is used.
  vendorHash = null;
  doCheck = true;
  patches = [ ./jdtls-wrapper.patch ];
  postPatch = ''
    cp ${./initialization-settings.go} initialization-settings.go
    cp ${../tests/jdtls-wrapper_test.go} main_test.go
    cp ${../tests/initialization-settings_test.go} initialization-settings_test.go
  '';
  preCheck = ''
    export TEST_SHELL=${runtimeShell}
  '';
  meta = {
    description = "JDTLS virtual-source bridge for Helix";
    homepage = "https://github.com/quantonganh/jdtls-wrapper";
    license = lib.licenses.mit;
    mainProgram = "jdtls-wrapper";
  };
}
