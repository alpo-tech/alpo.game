# Deployment

The project deploys to a VPS from GitHub Actions when changes are pushed or
merged into `main`.

## GitHub Secrets

Add these repository secrets:

- `VPS_HOST`: VPS hostname or IP address.
- `VPS_USER`: SSH user used by GitHub Actions.
- `VPS_SSH_KEY`: private SSH key for `VPS_USER`.
- `VPS_PORT`: optional SSH port. Defaults to `22`.

## VPS Requirements

The SSH user must be able to run `sudo` for these operations:

- create the `alpo-game` system user if it does not exist;
- write releases under `/opt/alpo-game`;
- install `/etc/systemd/system/alpo-game.service`;
- run `systemctl daemon-reload`, `enable`, `restart`, and `status`.

The service listens on `127.0.0.1:8081`/`:8081` from the Go process. Put a
reverse proxy such as nginx in front of it if the game should be available on a
domain or through HTTPS.

## Release Layout

GitHub Actions stores releases like this:

```text
/opt/alpo-game/
  current -> /opt/alpo-game/releases/<commit-sha>
  releases/
    <commit-sha>/
      alpo-game
      web/
```
