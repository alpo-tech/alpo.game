# alpo.game

`alpo.game` is a local browser-based Sea Battle game for two players. One
machine runs the Go server, and two browsers connect to it over `localhost` or
the local network.

## Features

- Two-player game in separate browser windows or on two computers in one local
  network.
- 10x10 board with the fleet `5, 4, 3, 3, 2`.
- Automatic ship placement in the browser.
- Server-side validation for ship size, board bounds, overlaps, and touching
  ships.
- Turn-based shooting: a hit keeps the turn, a miss passes it to the opponent.
- In-memory game state with a `New game` button for quick resets.

## Requirements

- Go 1.18 or newer.

## Run Locally

```bash
go run ./server
```

Open two browser windows at:

```text
http://localhost:8080
```

Join the game in each window, press `Auto place`, then `Ready`. When both
players are ready, the first player starts shooting at the enemy board.

## Play From Another Computer

Start the server on the host machine:

```bash
go run ./server
```

Find the host machine's local IP address. On macOS, this is usually:

```bash
ifconfig
```

Look for an active `inet` address on `en0`, for example `192.168.1.67`.
Players on the same Wi-Fi or LAN can then open:

```text
http://192.168.1.67:8080
```

If the page does not open from another computer, check that both devices are on
the same network and that the operating system firewall allows incoming
connections to port `8080`.

## Gameplay

1. Open the game in two browsers or on two computers.
2. Click `Join game` in both browsers.
3. Click `Auto place` to generate a valid fleet.
4. Click `Ready`.
5. Shoot at empty cells on the enemy board when it is your turn.
6. Sink every enemy ship to win.

Cell colors:

- Gray/green cells on your board are your ships.
- White dot means miss.
- Red means hit.
- Gold means a sunk ship.

## API

The browser uses a small JSON API:

- `POST /api/join` creates a player session.
- `GET /api/state?playerId=...` returns the current player view.
- `POST /api/place` submits a fleet.
- `POST /api/shoot` fires at a cell.
- `POST /api/reset` resets the in-memory game.

The server keeps all state in memory. Restarting the process clears the game.

## Development

Run tests:

```bash
go test ./...
```

Format Go code:

```bash
gofmt -w server/main.go app/model/model.go app/model/model_test.go
```

## Logging

The server writes structured JSON logs. Normal requests are logged at `info`,
client errors at `warn`, and server errors at `error`. The polling endpoint
`/api/state` is logged at `debug` to keep routine game updates from flooding the
console during normal runs.
