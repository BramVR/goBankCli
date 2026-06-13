---
summary: "Install gobankcli from source and verify the local binary."
read_when:
  - "Updating public install instructions."
  - "Checking the site-first setup path."
title: "Install"
description: "Build gobankcli from source, install the binary locally, and verify the command surface."
---
# Install

`gobankcli` currently installs from source. The binary is local and self-contained: no daemon, cloud service, browser extension, or background process is installed.

## Requirements

- Go toolchain available on your machine.
- Git access to the public source repository.
- A terminal where `~/.local/bin` or another user binary directory can be added to `PATH`.

## Build From Source

```bash
git clone https://github.com/BramVR/goBankCli.git
cd goBankCli
make build
./bin/gobankcli --help
```

The build writes `./bin/gobankcli`.

## Install For Your User

```bash
mkdir -p ~/.local/bin
install -m 755 ./bin/gobankcli ~/.local/bin/gobankcli
gobankcli --help
```

If `gobankcli` is not found, add `~/.local/bin` to your `PATH` in your shell profile.

## Run Without Installing

```bash
go run ./cmd/gobankcli --help
```

This is useful for local development or trying a branch before installing the binary.

## First Command

After installation, initialize local paths and inspect the setup:

```bash
gobankcli init
gobankcli doctor
```

`init` creates local config, database, and export directories. It does not store provider secrets in the config file. `doctor` reports whether paths exist and whether provider credential environment variables are present.

Next: [Quickstart](quickstart.md).
