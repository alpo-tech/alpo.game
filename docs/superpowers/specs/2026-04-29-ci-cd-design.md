# CI/CD Design

## Goal

Add GitHub Actions CI for every pull request and push to `main`, plus automatic
deployment to a VPS when `main` changes.

## Architecture

CI runs in GitHub-hosted Ubuntu runners and validates formatting, static checks,
tests, and server compilation. CD repeats the checks, cross-compiles a Linux
binary, uploads the binary plus `web/` assets to the VPS over SSH, atomically
updates `/opt/alpo-game/current`, and restarts the `alpo-game` systemd service.

## Components

- `.github/workflows/ci.yml`: validates code on PRs and pushes to `main`.
- `.github/workflows/deploy.yml`: deploys to the VPS on pushes to `main` and on
  manual `workflow_dispatch`.
- `deploy/alpo-game.service`: systemd unit that runs the Go binary from
  `/opt/alpo-game/current` on port `8081`.
- `deploy/README.md`: operator instructions for secrets and server
  prerequisites.

## Secrets

Deployment uses `VPS_HOST`, `VPS_USER`, `VPS_SSH_KEY`, and optional `VPS_PORT`.
The VPS user needs sudo rights for installing releases and restarting systemd.

## Testing

The workflow runs `gofmt`, `go vet ./...`, `go test ./...`, and `go build`.
The deploy workflow also performs a smoke test against `http://127.0.0.1:8081/`
on the VPS after restart.
