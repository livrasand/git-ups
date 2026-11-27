# Git Uninterruptible Power Supply

`git-ups` is the Go client for **gitGost**'s resilient Git networking layer (**GREN**).  
It wraps ordinary Git operations (`clone`, `fetch`, `pull`, `push`) in a persistent background job queue with retry, pause/resume, cancellation, and progress tracking. It installs itself as a `git` subcommand, so it is invoked as `git ups ...`.

> Status: Phase 0 / beta. The client is under active development; expect rough edges and breaking changes.

## Features

- **Background jobs** for `clone`, `fetch`, `pull`, and `push`, with a SQLite-backed queue.
- **Resumable downloads**: server-side bundle creation, HTTP `Range` downloads, SHA-256 verification, and automatic fallback to layered shallow cloning.
- **Retry with backoff** for transient network errors (DNS, connection reset, timeouts, 5xx, etc.).
- **Job control**: `pause`, `resume`, `cancel`, and live progress via `watch`.
- **URL rewriting** for GitHub, GitLab, and Codeberg repositories into the gitGost gateway.
- **Cross-platform**: Linux, macOS, and Windows (amd64/arm64; Windows amd64).

## Requirements

- Go 1.25+ (see `go.mod`)
- A working `git` installation on `PATH`

## Building

```sh
go build -o git-ups ./cmd
```

To embed a version string (as the release workflow does):

```sh
go build \
  -ldflags "-X github.com/livrasand/git-ups/internal/cli.version=1.0.0" \
  -o git-ups ./cmd
```

## Installing the client

### Quick install (macOS and Linux)

```sh
curl -L https://git-ups.livrasand.com/install.sh | sh
```

### Quick install (Windows)

```powershell
pwsh -Command "iwr https://git-ups.livrasand.com/install.ps1 -useb | iex"
```

The scripts download the latest release, verify its SHA-256 checksum and run `git-ups install`. Source: [`scripts/install.sh`](scripts/install.sh) and [`scripts/install.ps1`](scripts/install.ps1).

### With Homebrew (macOS)

```sh
brew tap livrasand/tap
brew install git-ups
```

This installs the `git ups` command.

To update later:

```sh
brew upgrade git-ups
```

### From a release binary or source

After building (or downloading a release binary), run:

```sh
./git-ups install
```

This creates the job database and copies the binary to `~/.local/bin/git-gost` and `~/.local/bin/git-ups` (adding it to your shell `PATH` when possible). Afterwards you can use:

```sh
git ups <command>
```

## Usage

```
git ups clone <url> [directory]   Clone via the gitGost server (background by default)
git ups fetch [args...]           Create a fetch job in the current repository
git ups pull [args...]            Create a pull job in the current repository
git ups push [args...]            Create a push job in the current repository
git ups jobs [n]                  List the last n jobs (default 50)
git ups watch <id>                Follow a job's progress
git ups pause <id>                Pause a job
git ups resume <id>               Resume a paused job
git ups cancel <id>               Cancel a job
git ups install                  Prepare the client environment
git ups version                  Show the version
```

Any other command is passed through to `git` unchanged.

Job options:

```
-f, --foreground   Run in the foreground with live progress
```

Example:

```sh
git ups clone https://github.com/torvalds/linux.git
# Job created.
# ID: 42
# Running in background...
git ups watch 42
```

## Environment variables

| Variable | Default | Description |
| --- | --- | --- |
| `GITGOST_SERVER` | `https://gitgost.livrasand.com` | Base URL of the gitGost gateway. |
| `GITGOST_HOME` | `~/.gitgost` | Data directory (SQLite queue and bundle cache). |
| `GITGOST_DEBUG` | unset | Enable debug output in Git helper routines when set to `1`. |

## Supported remotes

Repository URLs are rewritten for the following hosts:

- `github.com`
- `gitlab.com`
- `codeberg.org`

Both HTTPS and SCP-like SSH URLs (`git@host:owner/repo.git`) are accepted.

## How cloning works

1. The client asks the gateway to create a remote bundle job (`POST /v2/jobs`).
2. It polls the job until the bundle is ready, then downloads it with HTTP `Range` requests (resumable) into `~/.gitgost/cache/<id>.bundle`.
3. The bundle is verified against its SHA-256 hash and materialized with `git clone`.
4. The remote is re-pointed at the original repository URL.
5. If the gateway does not support remote jobs, the client falls back to a layered clone: shallow fetch in chunks, progressive `--deepen`, and a final `--unshallow`.

## Why a server is required

The client's resilience machinery (queue, retry, pause/resume, progress) is local, but **`clone` is designed around a gitGost gateway**:

- `git ups clone <url>` always rewrites the source URL to `GITGOST_SERVER` (GitHub/GitLab/Codeberg → gateway path).
- It asks the gateway to create a remote bundle job, polls it, downloads the resulting bundle over HTTP with `Range` requests, verifies its SHA-256, and materializes it locally.
- Even the layered fallback still fetches from the gateway URL, not directly from the original forge.

The gateway is what makes the clone resilient and (optionally) anonymous: it generates the bundle, distributes it, and anonymizes commit metadata. Without it, the client has no source for the bundle and no anonymous transport.

By contrast, `fetch`, `pull`, and `push` jobs simply run `git` against the remotes already configured in your repository:

- If your remote points at the gitGost gateway, traffic (and anonymization) goes through the server.
- If your remote points directly at GitHub/GitLab/Codeberg, those jobs work without any gitGost server, using only the local resilience layer.

So: **resilient fetch/pull/push can be used server-free; resilient clone cannot**, unless the client is modified to clone directly from the forge with retry logic.

## Job storage

Jobs are stored in a local SQLite database:

```
~/.gitgost/gitgost.db
```

Bundle downloads are cached under:

```
~/.gitgost/cache/
```

## Development

Run the test suite:

```sh
go test ./...
```

Project layout:

```
cmd/                 Entry point
internal/cli/        Command handling, URL rewriting, background process control
internal/jobs/       Job queue, store, retry logic, clone strategies
internal/git/        Git helpers (bundle creation, commit anonymization, push)
internal/config/     Configuration loading
internal/utils/      Logging and temporary directory helpers
```

## Privacy note

gitGost anonymizes commit metadata on the server side. Your IP address is still visible to the gitGost server; for stronger protection use Tor or a VPN when pushing or cloning through the gateway. See the upstream project's threat model for details.

## License

This project is licensed under the [MIT License](LICENSE).

[![Share](https://img.shields.io/badge/share-000000?logo=x&logoColor=white)](https://x.com/intent/tweet?text=Check%20out%20this%20project%20on%20GitHub:%20https://github.com/livrasand/git-ups)
[![Share](https://img.shields.io/badge/share-1877F2?logo=facebook&logoColor=white)](https://www.facebook.com/sharer/sharer.php?u=https://github.com/livrasand/git-ups)
[![Share](https://custom-icon-badges.demolab.com/badge/share-0A66C2?logo=linkedin-white&logoColor=white)](https://www.linkedin.com/sharing/share-offsite/?url=https://github.com/livrasand/git-ups)
[![Share](https://img.shields.io/badge/share-FF4500?logo=reddit&logoColor=white)](https://www.reddit.com/submit?title=Check%20out%20this%20project%20on%20GitHub:%20https://github.com/livrasand/git-ups)
[![Share](https://img.shields.io/badge/share-0088CC?logo=telegram&logoColor=white)](https://t.me/share/url?url=https://github.com/livrasand/git-ups&text=Check%20out%20git-ups:%20https://github.com/livrasand/git-ups)
[![Share](https://img.shields.io/badge/share-0285FF?logo=bluesky&logoColor=white)](https://bsky.app/intent/compose?text=Check%20out%20git-ups:%20https://github.com/livrasand/git-ups)
[![Share](https://img.shields.io/badge/share-6364FF?logo=mastodon&logoColor=white)](https://mastodonshare.com/?url=https://github.com/livrasand/git-ups)
[![Share](https://img.shields.io/badge/share-FF6600?logo=ycombinator&logoColor=white)](https://news.ycombinator.com/submitlink?u=https://github.com/livrasand/git-ups)
[![Share](https://img.shields.io/badge/share-000000?logo=threads&logoColor=white)](https://www.threads.net/intent/post?text=Check%20out%20git-ups:%20https://github.com/livrasand/git-ups)
[![Share](https://img.shields.io/badge/share-25D366?logo=whatsapp&logoColor=white)](https://wa.me/?text=Check%20out%20git-ups:%20https://github.com/livrasand/git-ups)

*✨ Thanks for visiting **Git UPS**!*

<img src="https://visitor-badge.laobi.icu/badge?page_id=livrasand.git-ups&style=for-the-badge&color=00d4ff" alt="Views">
