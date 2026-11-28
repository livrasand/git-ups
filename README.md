# Git UPS

<p align="center">
    <img src="git-ups-workflow.gif" width="800" alt="Git UPS UI Demo">
</p>

**Reliable Git transfers that survive network interruptions.**

Git UPS is an open-source, cross-platform Git client that makes `clone`, `fetch`, `pull` and `push` more resilient with background jobs, automatic retries and resumable transfers.

**Your connection can drop. Your Git operations don't have to.**

[Website](https://git-ups.livrasand.com/) · [GitHub](https://github.com/livrasand/git-ups) · [Report a bug](https://github.com/livrasand/git-ups/issues)

## Features

- **Automatic retries** — Recover from temporary network failures.
- **Background jobs** — Run Git operations without blocking your terminal.
- **Resumable cloning** — Resume interrupted downloads.
- **Job control** — Pause, resume, cancel and monitor transfers.
- **Persistent storage** — Keep job state across restarts with SQLite.
- **Cross-platform** — Linux, macOS and Windows.
- **Open source** — MIT licensed, with no accounts or telemetry.

## Installation


<img alt="Shell running the brew installation" width="500" src="git-ups-install.gif">

### macOS / Linux

    curl -fsSL https://git-ups.livrasand.com/install.sh | sh

### Windows

    pwsh -Command "iwr https://git-ups.livrasand.com/install.ps1 -useb | iex"

### Homebrew

    brew tap livrasand/tap
    brew install git-ups

## Usage

Run familiar Git commands through `git ups`:

    git ups clone https://github.com/torvalds/linux.git
    git ups fetch
    git ups pull
    git ups push

Manage background jobs:

    git ups jobs
    git ups watch 42
    git ups pause 42
    git ups resume 42
    git ups cancel 42

Use `--foreground` to monitor operations directly in your terminal.

## Requirements

- Go 1.25+ (for building from source).
- Git installed and available in `PATH`.

## How it works

Git UPS runs Git operations through a persistent local job queue, automatically retries recoverable failures and tracks transfer progress.

Job data is stored locally in `~/.gitgost/`.

**Note:** `clone` currently uses the [gitGost](https://github.com/livrasand/gitGost) gateway. `fetch`, `pull` and `push` also support direct Git remotes.

---

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
