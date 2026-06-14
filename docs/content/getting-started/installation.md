---
title: "Installation"
description: "Install openf1 from a release, with go install, or from source."
weight: 20
---

## Prebuilt binaries

Every [release](https://github.com/tamnd/openf1-cli/releases) carries archives for Linux, macOS,
and Windows on amd64 and arm64, plus deb, rpm, and apk packages for Linux.
Download, unpack, put `openf1` on your `PATH`, done. The `checksums.txt`
on each release is signed with keyless [cosign](https://docs.sigstore.dev/) if
you want to verify before running.

## With Go

```bash
go install github.com/tamnd/openf1-cli/cmd/openf1@latest
```

That puts `openf1` in `$(go env GOPATH)/bin`, which is `~/go/bin` unless
you moved it. Make sure that directory is on your `PATH`.

## From source

```bash
git clone https://github.com/tamnd/openf1-cli
cd openf1-cli
make build        # produces ./bin/openf1
./bin/openf1 version
```

## Container image

```bash
docker run --rm ghcr.io/tamnd/openf1:latest --help
```

## Checking the install

```bash
openf1 version
```

prints the version and exits.
