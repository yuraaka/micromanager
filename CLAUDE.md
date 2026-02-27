# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

MicroManager (`mm`) is a CLI tool for scaffolding and managing microservice repositories. It uses language packs with Go templates to generate service boilerplate (HTTP servers, clients, configs, Dockerfiles). Currently supports Go services only.

## Build & Run

```bash
# Build the CLI binary
go build -o mm ./cmd/mm

# Run tests (no test files exist yet — standard Go testing)
go test ./...

# Vet/lint
go vet ./...
```

## Architecture

**CLI entry point**: `cmd/mm/main.go` — all Cobra commands defined in a single file.

**Core packages** (all under `internal/`):
- `config` — loads/saves repo defaults (`.mm/defaults.toml`) and per-service config (`services/<name>/service.toml`), both TOML-based
- `lang` — language pack discovery, validation, and template rendering engine. Packs are searched first in `.mm/packs/`, then in `pack/lang/<language>/` walking up to 5 parent directories
- `scaffold` — repo initialization (`mm init`) and service creation (`mm new`). Delegates to `lang.ApplyService` when a pack is found
- `runtime` — builds and runs services with environment variables loaded from `service.toml` by mode (local/docker/minikube)
- `testing` — test orchestration (stub)

**Language packs** (`pack/lang/go/`):
- `templates/service/*` → copied to `services/<name>/`
- `templates/common/*` → copied to `common/` (full overwrite each time)
- `templates/root/*` → copied to repo root
- Files ending in `.tmpl` are rendered through `text/template` then the suffix is stripped
- Template variables: `{{.ProjectName}}`, `{{.ServiceName}}`
- Template functions: `snake`, `kebab`, `camel`, `title`, `upper`, `lower`, `joinPath`

**Generated service pattern** (Go pack): Gin HTTP server, service interface + implementation, HTTP client, structured logging (logrus), env-based config, multi-stage Docker build.
