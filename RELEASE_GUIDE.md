Release Guide

This repo ships prebuilt binaries via GoReleaser and an npm wrapper (`npm/`) that downloads the correct binary on install. Follow this guide to cut and publish a new version.

Prerequisites
- `gh` authenticated: `gh auth login` (used by `mise run release_gh`).
- `GITHUB_TOKEN` available to GoReleaser (the `release_gh` task reads it from `gh auth token`).
- Node/npm installed for publishing the npm package.
- Ensure `.goreleaser.yml` archives produce binary-only tarballs (no extra files).

Versioning Rules
- Git tag: `vX.Y.Z` (with leading `v`).
- npm package version (`npm/package.json`): `X.Y.Z` (no `v`).
- Asset names (from GoReleaser): `copilot-proxy_{{.Version}}_{{.Os}}_{{.Arch}}.tar.gz`.

Standard Release Workflow
1) Pick a version `X.Y.Z` (SemVer).
2) Update npm version to match:
   - Edit `npm/package.json` to set `"version": "X.Y.Z"` (or run `npm version X.Y.Z --no-git-tag-version` in `npm/`).
3) Commit the change:
   - `git add npm/package.json && git commit -m "Release vX.Y.Z"`
4) Create and push tag:
   - `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
   - `git push && git push --tags`
5) Create GitHub draft release artifacts:
   - `mise run release_gh` (runs GoReleaser with `--clean`, producing a Draft release on GitHub).
6) Verify assets in the Draft release:
   - Assets should be `copilot-proxy_X.Y.Z_<os>_<arch>.tar.gz` plus `checksums.txt`.
   - Spot check one asset: download locally and confirm `tar -tzf` contains only the `copilot-api-proxy[.exe]` binary.
7) Publish the GitHub release (remove Draft):
   - Npm postinstall downloads from the published `releases/download/vX.Y.Z/…` URL; Draft releases will 404.
8) Publish the npm package:
   - `cd npm`
   - First-time only: `npm publish --access public`
   - Subsequent releases: `npm publish`
   - `.npmignore` ensures local binaries aren’t bundled.
9) Sanity check install:
   - `npm install -g copilot-api-proxy@X.Y.Z`
   - Confirm the install downloads the right asset and places the `copilot-api-proxy` binary in the package directory.

Local Dry‑Run (optional)
- Build snapshot artifacts locally: `mise run release`
- Test npm install against local `dist`: `COPILOT_PROXY_BASE_URL="file://$(pwd)/dist" npm install -g ./npm`

Troubleshooting
- 404 during npm postinstall: The GitHub release is likely still a Draft or versions don’t match; publish the release and ensure `npm/package.json` version matches the tag.
- Checksum mismatch: Rebuild and re-upload assets; don’t edit tarballs after checksums are generated.
- Wrong asset name: Ensure `.goreleaser.yml` `archives.name_template` is `copilot-proxy_{{ .Version }}_{{ .Os }}_{{ .Arch }}`.

Handy Commands
- Format/build locally: `mise run fmt && mise run build`
- Release (GitHub Draft): `mise run release_gh`
- Snapshot release (local `dist/`): `mise run release`

