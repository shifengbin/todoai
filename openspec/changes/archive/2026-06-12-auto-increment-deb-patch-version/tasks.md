## 1. Version State

- [x] 1.1 Add a root-level `VERSION` file seeded from the current deb package version.
- [x] 1.2 Update `scripts/package-deb.sh` to read and validate three-part numeric semver from `VERSION` when no `VERSION` override is provided.
- [x] 1.3 Add patch increment logic so the default path uses the next patch version for package metadata and output filename.

## 2. Packaging Flow

- [x] 2.1 Update the script to persist the package version only after Wails build, package assembly, and `dpkg-deb` complete successfully.
- [x] 2.2 Preserve explicit `VERSION=... scripts/package-deb.sh` overrides and persist the explicit version after successful packaging.
- [x] 2.3 Ensure script errors for missing or invalid version input are clear and do not modify the persisted version.
- [x] 2.4 Update README Debian packaging documentation so it describes automatic patch incrementing without relying on a fixed patch example.

## 3. Verification

- [x] 3.1 Add or run focused shell-level checks for patch increment, explicit override, invalid version rejection, and failure preserving the previous version.
- [x] 3.2 Run client automated tests with `npm test` in `frontend/`.
- [x] 3.3 Run an automated review command such as `git diff --check` and resolve any formatting issues it reports.
- [x] 3.4 Run `openspec status --change auto-increment-deb-patch-version`.
- [x] 3.5 Run `wails build -tags webkit2_41` to generate the executable file.
