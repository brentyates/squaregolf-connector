# Release Process

## Current release conventions

- Stable releases use tags like `v0.1.1`.
- Pre-releases use tags like `v0.1.0-alpha.11`.
- Pushing a tag that matches `v*` triggers two GitHub Actions workflows in parallel: `.github/workflows/release-macos.yml` and `.github/workflows/release-windows.yml`. Both upload their archive to the same GitHub release.
- Tags containing `-` are published as GitHub prereleases. Tags without `-` are published as normal releases.

## Version files to update

Before creating a release, update:

- `internal/version/version.go`
  - Set `Version` to the release version, for example `0.1.1`.
- `macos/Info.plist`
  - Set `CFBundleShortVersionString` to the app version, for example `0.1.1`.
  - Increment `CFBundleVersion` for each shipped macOS build.

The Windows resource file (`windows/app.rc`) is templated at build time with the tag passed into the workflow, so no manual edits are required there.

## Release steps

1. Make and verify the intended code changes.
2. Run `go test ./...`.
3. Update release version values in:
   - `internal/version/version.go`
   - `macos/Info.plist`
4. Commit the changes to `main`.
5. Create an annotated tag for the release:
   - Stable: `git tag -a v0.1.1 -m "Release v0.1.1"`
   - Pre-release: `git tag -a v0.1.0-alpha.12 -m "Release v0.1.0-alpha.12"`
6. Push the branch and tag:
   - `git push origin main`
   - `git push origin v0.1.1`
7. Confirm the GitHub Actions release workflows start:
   - `gh run list --workflow "release-macos.yml" --limit 5`
   - `gh run list --workflow "release-windows.yml" --limit 5`
8. Confirm the GitHub release and both uploaded archives (macOS and Windows) after the workflows complete.

## Notes

- The macOS release archive is built by `scripts/package-macos-release.sh`.
- The Windows release archive is built by `scripts/package-windows-release.sh`, which calls `scripts/build-windows-app.sh` (requires MSYS2 + MinGW on the host).
- Both release workflows use the tag name as the published GitHub release title.
- If a release should be stable, do not use a hyphenated prerelease tag.
