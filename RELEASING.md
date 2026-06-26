# Releasing todo-ui

`scripts/release.sh` is the whole release cycle in one command: it picks the
version, runs the tests, signs + notarizes the macOS binaries, builds the
Linux/Windows archives, writes `SHA256SUMS`, and — after you confirm — tags,
pushes, and creates the GitHub release with every artifact attached.

For the one-time Apple cert / notary setup it depends on, see
[`SIGNING.md`](SIGNING.md).

## TL;DR

```bash
scripts/release.sh
```

That's it. With `SIGN_IDENTITY` exported in your shell (see below) and no version
given, it auto-selects the next version, builds everything, shows you what it
made, and asks once before anything leaves your machine.

## One-time setup on this Mac

1. Apple Developer ID cert + notary profile exist — see `SIGNING.md`.
2. `SIGN_IDENTITY` is exported in your shell (e.g. `~/.zshrc`):

   ```bash
   export SIGN_IDENTITY="Developer ID Application: <YOUR NAME> (<TEAMID>)"
   ```

   > Use the exact name of your Developer ID Application certificate. List the
   > ones in your keychain with:
   >
   > ```bash
   > security find-identity -v -p codesigning
   > ```
   >
   > If more than one Developer ID Application cert exists, the script won't
   > auto-guess — the export tells it which to use. After editing your shell rc,
   > open a new terminal (or `source` it).

3. `gh` is authenticated (`gh auth status`).

## Versioning: snapshot (`-dev`) between releases

This repo follows a Maven-style snapshot flow:

- Between releases, `var version` in `main.go` carries a **`-dev`** suffix for
  the *upcoming* version — e.g. after `v0.2.1` ships, `main.go` reads
  `v0.2.2-dev`. Local/dev builds therefore self-identify as `v0.2.2-dev`
  (`todo-ui --version`), clearly "newer than 0.2.1, not yet 0.2.2". The suffix
  also sorts *below* the clean version in SemVer, mirroring Maven's SNAPSHOT.
- A release **strips the `-dev`** to a clean `vX.Y.Z`, commits that, then
  **tags and releases the clean version** (the git tag and GitHub release never
  carry `-dev`).
- Right after publishing, the script **bumps `var version` to the next
  `-dev` snapshot** (`v0.2.3-dev`), commits `Start v0.2.3-dev development`, and
  pushes — opening the next dev cycle automatically. Don't edit `var version`
  by hand.

## How the version is chosen

You normally don't pass a version. The script infers it:

- It starts from the clean release form of `main.go`'s `var version` (so
  `v0.2.2-dev` → `v0.2.2`), floored at the highest **pushed** tag / published
  release, and skips anything already tagged or released.
- With `main.go` at `v0.2.2-dev` and `v0.2.1` released, the next release is
  **`v0.2.2`**; afterwards `main.go` becomes `v0.2.3-dev` and the following run
  yields `v0.2.3`.

To force a specific version, pass it explicitly:

```bash
scripts/release.sh 0.2.0      # the leading "v" is optional
scripts/release.sh v0.2.0     # same thing
```

The version must be **`X.Y.Z`** — three numeric parts (e.g. `0.1.6`, `1.2.10`).
The `v` prefix is added for you if you omit it; anything that isn't `X.Y.Z` is
rejected. An explicit version is also refused if it's already tagged or released.

> A stray **local-only** `v0.2.0` tag exists in this repo (never pushed/released).
> The auto-picker ignores unpushed tags when choosing the next version, but it
> will never *reuse* one — so it can't accidentally clobber it. The script prints
> a note when a higher local tag exists. Delete it with `git tag -d v0.2.0` if
> it's cruft.

## What a full run does

```
scripts/release.sh
  ├─ preflight: clean tree · gh auth · fetch tags
  ├─ resolve clean release version (strips main.go's -dev); bump+commit if needed
  ├─ go test ./...
  ├─ Todoist integration guard (live API; runs if a token is present)
  ├─ build into dist/:
  │     macOS arm64 + amd64  → sign → notarize (uploads to Apple)
  │     linux amd64 + arm64  → .tar.gz
  │     windows amd64        → .zip
  │     SHA256SUMS.txt
  ├─ print artifact list
  └─ "Proceed? [y/N]"   ← nothing pushed before this
        ├─ git tag (clean vX.Y.Z) · push branch + tag · gh release create --generate-notes
        ├─ rebuild ./todo-ui at the released version (so symlinked local
        │  installs, e.g. /opt/homebrew/bin/todo-ui, run it immediately)
        └─ bump main.go to the next -dev snapshot · commit · push
```

## Modes & flags

| Command | What it does |
|---|---|
| `scripts/release.sh` | Auto-version, full build + sign + notarize, prompt, then publish. |
| `scripts/release.sh v0.2.0` | Same, but for an explicit version. |
| `scripts/release.sh --dry-run` | **Validate only** — tests, compile every target, print the plan; changes nothing (no bump/commit/tag/push/notarize/`dist`). Nothing to undo. |
| `scripts/release.sh --tag-only` | **Just create & push the version tag — no build, no release.** |
| `scripts/release.sh --tag-only --no-publish` | Create the tag **locally only** (don't push). |
| `scripts/release.sh --no-publish` | Build everything into `dist/`, print the manual publish commands, push nothing. |
| `scripts/release.sh --yes` | Skip the confirmation prompt (unattended). |
| `scripts/release.sh --skip-mac` | Skip macOS sign/notarize; build Linux/Windows only (e.g. before the cert exists). |
| `scripts/release.sh --skip-tests` | Skip the `go test ./...` gate (also skips the integration guard). |

Flags combine, e.g. `scripts/release.sh v0.1.7 --skip-mac --no-publish`.

### Todoist integration guard

`go test ./...` excludes the live-API tests (they're behind the `integration`
build tag). The release runs them separately **when a Todoist token is
available** — they hit the real server to confirm the endpoints todo-ui relies
on (token validation, full sync, filter, **completed-tasks fetch**, and the
item add/complete/uncomplete/delete commands) still work, since that's the part
most likely to break.

```bash
scripts/todoist-api-test.sh                      # read-only checks
TODOUI_INTEGRATION_WRITE=1 scripts/todoist-api-test.sh   # + create→complete→reopen→delete round-trip
```

Set `SKIP_INTEGRATION=1` to skip it during a release even when a token is
present. Without any token the guard is skipped (never a release blocker).

### `--tag-only`

Use this when you want to *mark* a version without producing or publishing
binaries — e.g. to record a tag now and run the signed build later. It runs
preflight, resolves/bumps the version, creates an annotated tag, and pushes it
(branch + tag). Add `--no-publish` to keep the tag local. No artifacts are built
or uploaded.

## Overrides

| Env var | Default | Purpose |
|---|---|---|
| `SIGN_IDENTITY` | (must be set if multiple certs exist) | The Developer ID Application identity to sign with. |
| `NOTARY_PROFILE` | `todoui-notary` | The `notarytool` keychain profile. |

## Safety properties

- **Won't run on a dirty tree** — so the tag and version-bump commit are meaningful.
- **Version ↔ source stay in sync** — the binary's `var version` always matches the tag.
- **No accidental publish** — pushing only happens at the confirmation (or `--yes`).
- **Won't reuse a version** — aborts if the chosen tag/release already exists.
- **Won't sign with the wrong cert** — refuses to guess when multiple certs exist.

## Verifying a published artifact

```bash
unzip todo-ui_v0.1.6_darwin_arm64.zip
spctl -a -vvv -t install ./todo-ui_v0.1.6_darwin_arm64/todo-ui   # should say: accepted
shasum -a 256 -c SHA256SUMS.txt                                  # match the release
```
