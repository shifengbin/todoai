#!/usr/bin/env bash
set -euo pipefail

APP_NAME="tui-helper"
VERSION="${VERSION:-0.1.2}"
ARCH="${ARCH:-amd64}"
MAINTAINER="${MAINTAINER:-FengbinShi <shifengbin@jiandan100.cn>}"
DESCRIPTION="Desktop project shell helper"
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PACKAGE_DIR="${ROOT_DIR}/build/deb/${APP_NAME}_${VERSION}_${ARCH}"
OUTPUT_DEB="${ROOT_DIR}/build/bin/${APP_NAME}_${VERSION}_${ARCH}.deb"

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
Version: ${VERSION}
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
Name=TUI Helper
Comment=${DESCRIPTION}
Exec=/usr/bin/${APP_NAME}
Icon=${APP_NAME}
Terminal=false
Categories=Development;Utility;
EOF
chmod 644 "${PACKAGE_DIR}/usr/share/applications/${APP_NAME}.desktop"

dpkg-deb --build --root-owner-group "${PACKAGE_DIR}" "${OUTPUT_DEB}"
echo "${OUTPUT_DEB}"
