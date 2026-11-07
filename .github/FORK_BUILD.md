# Fork image build

The fork keeps two independent patches based on upstream v2.6.1
(`a0b1cf58a39b2fdede2cf427bd7c782bb90d511a`):

| Branch | Purpose |
| --- | --- |
| `fix/explorer-request-headers` | Bounded 64 KiB HTTP parser buffer for Explorer URL/Referer 431 failures |
| `fix/bounded-view-activity` | Explicit opt-in for bounded ClickHouse View activity |
| `builds/homelab` | Both patches and this fork's image pipeline |

Issue drafts are in [upstream-issues](upstream-issues). Publish and discuss those
with upstream before opening the two pull requests. The Dockerfile, workflow
and fork documentation belong only to `builds/homelab`; omit them from upstream
PRs. `main` continues to track upstream.

## Publication

`.github/workflows/homelab-image.yml` runs on pushes to `builds/homelab` in
`nik-makarov/logchef`. It requires native unit/race tests, vet, pinned lint and
sqlc checks before image publication. Native Ubuntu amd64/arm64 jobs build
`Dockerfile.homelab`; a final job publishes the combined OCI manifest to:

```text
ghcr.io/nik-makarov/logchef:sha-<full-commit>-<run-number>-<run-attempt>
ghcr.io/nik-makarov/logchef:homelab
```

The first tag identifies a single CI run. `homelab` moves after both architecture
builds succeed. Each architecture also has a run-specific `-amd64` or `-arm64`
tag. The job summary reports the reference; the manifest inspection reports its
digest. Use the manifest digest in deployment manifests for reproducibility.
CI generates the version/tag; local version bumps and image publication are
not part of this workflow. Rebuild by pushing a new commit to the build branch.

Publication uses the repository's `GITHUB_TOKEN` with `packages: write` only in
publication jobs. No personal token or Docker Hub credentials are required.
GitHub Actions must be enabled for the fork and its token allowed to publish
packages. After the first successful publication, check package visibility:
make the package public for anonymous pulls, or configure a read-only GHCR
image-pull secret in the consuming cluster. A public source repo alone does not
guarantee that a newly created package is public.

## Container contents

The frontend stage uses Bun Debian; the Go stage uses Bookworm; the runtime is
the pinned distroless static Debian non-root image. The workflow explicitly uses
Docker's bundled `docker` driver on native Ubuntu runners, so it does not pull
the default Alpine-based BuildKit daemon image. No QEMU/binfmt image is needed.
External Dockerfile frontend images are not used. Attestation and external
cache exports are disabled because the bundled driver may use the classic
Docker image store.
See Docker's [bundled driver documentation](https://docs.docker.com/build/builders/drivers/docker/).

Upstream `Dockerfile` and `Dockerfile.goreleaser` still contain Alpine. They
remain unchanged for upstream compatibility and are not used by this workflow.
The upstream release workflow also remains unchanged; this fork workflow
creates no Git tag or GitHub release.

The image contains the compiled frontend and Logchef at `/opt/logchef/logchef`,
runs as `65532:65532`, uses writable `/data` for the default SQLite database,
and accepts `-config /etc/logchef/config.toml`. Existing configuration can be
mounted at that path. It contains the Logchef application; any separately
deployed companion service needs its own image reference.

For an opted-in view, set `LOGCHEF_BOUNDED_ACTIVITY_VIEWS` to an exact
comma-separated list such as `default.logs_visible`. Empty means the original
sort-key safeguard remains in effect. Opted-in queries retain the 24-hour
window and hard scan/time budgets. No production database or deployment is
changed by preparing these branches.
