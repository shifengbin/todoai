#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

assert_eq() {
  local expected="$1"
  local actual="$2"
  local message="$3"

  if [[ "${actual}" != "${expected}" ]]; then
    fail "${message}: expected '${expected}', got '${actual}'"
  fi
}

assert_contains() {
  local haystack="$1"
  local needle="$2"
  local message="$3"

  if [[ "${haystack}" != *"${needle}"* ]]; then
    fail "${message}: expected to contain '${needle}', got '${haystack}'"
  fi
}

setup_fixture() {
  TEST_ROOT="$(mktemp -d)"
  export TEST_ROOT
  mkdir -p "${TEST_ROOT}/scripts" "${TEST_ROOT}/fake-bin"
  cp "${REPO_ROOT}/scripts/package-deb.sh" "${TEST_ROOT}/scripts/package-deb.sh"
  chmod +x "${TEST_ROOT}/scripts/package-deb.sh"

  cat > "${TEST_ROOT}/fake-bin/wails" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
mkdir -p build/bin build
printf '#!/usr/bin/env bash\n' > build/bin/tui-helper
chmod +x build/bin/tui-helper
printf 'fake icon\n' > build/appicon.png
EOF
  chmod +x "${TEST_ROOT}/fake-bin/wails"

  cat > "${TEST_ROOT}/fake-bin/dpkg-deb" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${DPKG_DEB_FAIL:-}" == "1" ]]; then
  echo "forced dpkg-deb failure" >&2
  exit 42
fi
output="${@: -1}"
mkdir -p "$(dirname "${output}")"
printf 'fake deb\n' > "${output}"
EOF
  chmod +x "${TEST_ROOT}/fake-bin/dpkg-deb"

  export PATH="${TEST_ROOT}/fake-bin:${PATH}"
}

teardown_fixture() {
  rm -rf "${TEST_ROOT}"
}

run_package_deb() {
  local stdout_file="${TEST_ROOT}/stdout"
  local stderr_file="${TEST_ROOT}/stderr"

  (cd "${TEST_ROOT}" && "$@" ./scripts/package-deb.sh >"${stdout_file}" 2>"${stderr_file}")
}

test_default_packaging_increments_patch_version() {
  setup_fixture
  trap teardown_fixture RETURN
  printf '0.1.8\n' > "${TEST_ROOT}/VERSION"

  run_package_deb env

  assert_eq "0.1.9" "$(cat "${TEST_ROOT}/VERSION")" "persisted version"
  assert_contains "$(cat "${TEST_ROOT}/stdout")" "build/bin/tui-helper_0.1.9_amd64.deb" "output path"
  assert_eq "Version: 0.1.9" "$(grep '^Version:' "${TEST_ROOT}/build/deb/tui-helper_0.1.9_amd64/DEBIAN/control")" "control version"
  [[ -f "${TEST_ROOT}/build/bin/tui-helper_0.1.9_amd64.deb" ]] || fail "expected deb artifact"
}

test_explicit_version_override_is_persisted() {
  setup_fixture
  trap teardown_fixture RETURN
  printf '0.1.9\n' > "${TEST_ROOT}/VERSION"

  run_package_deb env VERSION=0.2.0

  assert_eq "0.2.0" "$(cat "${TEST_ROOT}/VERSION")" "persisted explicit version"
  assert_contains "$(cat "${TEST_ROOT}/stdout")" "build/bin/tui-helper_0.2.0_amd64.deb" "explicit output path"
  assert_eq "Version: 0.2.0" "$(grep '^Version:' "${TEST_ROOT}/build/deb/tui-helper_0.2.0_amd64/DEBIAN/control")" "explicit control version"
}

test_invalid_persisted_version_is_rejected() {
  setup_fixture
  trap teardown_fixture RETURN
  printf '0.1\n' > "${TEST_ROOT}/VERSION"

  if run_package_deb env; then
    fail "expected invalid persisted version to fail"
  fi

  assert_eq "0.1" "$(cat "${TEST_ROOT}/VERSION")" "invalid version remains unchanged"
  assert_contains "$(cat "${TEST_ROOT}/stderr")" "Invalid package version" "invalid version error"
}

test_failed_packaging_preserves_previous_version() {
  setup_fixture
  trap teardown_fixture RETURN
  printf '0.1.8\n' > "${TEST_ROOT}/VERSION"

  if run_package_deb env DPKG_DEB_FAIL=1; then
    fail "expected forced dpkg-deb failure"
  fi

  assert_eq "0.1.8" "$(cat "${TEST_ROOT}/VERSION")" "failed packaging preserves version"
  assert_contains "$(cat "${TEST_ROOT}/stderr")" "forced dpkg-deb failure" "forced failure error"
}

test_default_packaging_increments_patch_version
test_explicit_version_override_is_persisted
test_invalid_persisted_version_is_rejected
test_failed_packaging_preserves_previous_version

echo "package-deb tests passed"
