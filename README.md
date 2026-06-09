# TUI Helper

## About

TUI Helper is a Wails desktop application with a Vue frontend and Go backend. It keeps a persisted list of local project directories and provides an embedded shell session for the selected project.

Package metadata:

- Wails major version: v2
- App name: TUI Helper
- Package name: `tui-helper`
- Version: `0.1.0`
- Maintainer: `FengbinShi <shifengbin@jiandan100.cn>`
- Icon: `build/appicon.png`

## Live Development

To run in live development mode, run `wails dev` in the project directory. This will run a Vite development
server that will provide very fast hot reload of your frontend changes. If you want to develop in a browser
and have access to your Go methods, there is also a dev server that runs on http://localhost:34115. Connect
to this in your browser, and you can call your Go code from devtools.

## Building

Backend tests:

```bash
go test ./...
```

Frontend tests and build:

```bash
cd frontend
npm run test
npm run build
```

Wails build:

```bash
wails build -tags webkit2_41
```

Debian package:

```bash
scripts/package-deb.sh
```

The Debian package script builds the Linux Wails binary, assembles package metadata, and writes `build/bin/tui-helper_0.1.0_amd64.deb`.
