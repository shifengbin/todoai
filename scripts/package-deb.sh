#!/usr/bin/env bash
set -euo pipefail

APP_NAME="todoai"
ARCH="${ARCH:-amd64}"
MAINTAINER="${MAINTAINER:-FengbinShi <shifengbin@jiandan100.cn>}"
DESCRIPTION="Desktop project shell helper"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION_FILE="${ROOT_DIR}/VERSION"

validate_version() {
  local version="$1"

  if [[ ! "${version}" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo "Invalid package version '${version}'. Expected numeric X.Y.Z semver." >&2
    exit 1
  fi
}

read_persisted_version() {
  if [[ ! -f "${VERSION_FILE}" ]]; then
    echo "Missing package version file: ${VERSION_FILE}" >&2
    exit 1
  fi

  local version
  version="$(<"${VERSION_FILE}")"
  validate_version "${version}"
  echo "${version}"
}

increment_patch_version() {
  local version="$1"
  local major minor patch

  validate_version "${version}"
  IFS="." read -r major minor patch <<< "${version}"
  echo "${major}.${minor}.$((10#${patch} + 1))"
}

if [[ -n "${VERSION+x}" ]]; then
  validate_version "${VERSION}"
  PACKAGE_VERSION="${VERSION}"
else
  PACKAGE_VERSION="$(increment_patch_version "$(read_persisted_version)")"
fi

PACKAGE_DIR="${ROOT_DIR}/build/deb/${APP_NAME}_${PACKAGE_VERSION}_${ARCH}"
OUTPUT_DEB="${ROOT_DIR}/build/bin/${APP_NAME}_${PACKAGE_VERSION}_${ARCH}.deb"

cd "${ROOT_DIR}"

wails build -clean -platform "linux/${ARCH}" -tags webkit2_41

if [[ ! -x "${ROOT_DIR}/build/bin/${APP_NAME}" ]]; then
  echo "Missing Wails binary: build/bin/${APP_NAME}" >&2
  exit 1
fi

rm -rf "${PACKAGE_DIR}"
install -Dm755 "${ROOT_DIR}/build/bin/${APP_NAME}" "${PACKAGE_DIR}/usr/bin/${APP_NAME}"
install -Dm644 "${ROOT_DIR}/build/appicon.png" "${PACKAGE_DIR}/usr/share/icons/hicolor/256x256/apps/${APP_NAME}.png"

install -d -m 755 "${PACKAGE_DIR}/DEBIAN" "${PACKAGE_DIR}/usr/share/applications"

cat > "${PACKAGE_DIR}/DEBIAN/control" <<EOF
Package: ${APP_NAME}
Version: ${PACKAGE_VERSION}
Section: utils
Priority: optional
Architecture: ${ARCH}
Maintainer: ${MAINTAINER}
Description: ${DESCRIPTION}
 Embedded desktop shell for local project directories.
EOF
chmod 644 "${PACKAGE_DIR}/DEBIAN/control"

cat > "${PACKAGE_DIR}/usr/share/applications/${APP_NAME}.desktop" <<EOF
[Desktop Entry]
Type=Application
Name=TodoAI
Comment=${DESCRIPTION}
Exec=/usr/bin/${APP_NAME}
Icon=${APP_NAME}
Terminal=false
Categories=Development;Utility;
EOF
chmod 644 "${PACKAGE_DIR}/usr/share/applications/${APP_NAME}.desktop"

dpkg-deb --build --root-owner-group "${PACKAGE_DIR}" "${OUTPUT_DEB}"
printf '%s\n' "${PACKAGE_VERSION}" > "${VERSION_FILE}"
echo "${OUTPUT_DEB}"
