const boardSize = 10;
const fleet = [5, 4, 3, 3, 2];
const storageKey = "alpo-player-id";

const state = {
  playerId: localStorage.getItem(storageKey) || "",
  playerNumber: 0,
  view: null,
  draftShips: [],
};

const statusEl = document.querySelector("#status");
const ownBoardEl = document.querySelector("#ownBoard");
const enemyBoardEl = document.querySelector("#enemyBoard");
const joinPanel = document.querySelector("#joinPanel");
const nameInput = document.querySelector("#nameInput");
const joinButton = document.querySelector("#joinButton");
const autoPlaceButton = document.querySelector("#autoPlaceButton");
const readyButton = document.querySelector("#readyButton");
const resetButton = document.querySelector("#resetButton");
const playerBadge = document.querySelector("#playerBadge");
const turnBadge = document.querySelector("#turnBadge");

joinButton.addEventListener("click", joinGame);
autoPlaceButton.addEventListener("click", () => {
  state.draftShips = generateFleet();
  render();
});
readyButton.addEventListener("click", placeFleet);
resetButton.addEventListener("click", resetGame);

renderEmptyBoards();
if (state.playerId) {
  pollState();
} else {
  render();
}
setInterval(() => {
  if (state.playerId) {
    pollState();
  }
}, 1000);

async function joinGame() {
  try {
    const data = await api("/api/join", {
      method: "POST",
      body: JSON.stringify({ name: nameInput.value.trim() }),
    });
    state.playerId = data.playerId;
    state.playerNumber = data.playerNumber;
    localStorage.setItem(storageKey, state.playerId);
    state.draftShips = generateFleet();
    await pollState();
  } catch (error) {
    setStatus(error.message);
  }
}

async function pollState() {
  try {
    state.view = await api(`/api/state?playerId=${encodeURIComponent(state.playerId)}`);
    state.playerNumber = state.view.playerNumber;
    render();
  } catch (error) {
    localStorage.removeItem(storageKey);
    state.playerId = "";
    state.playerNumber = 0;
    state.view = null;
    render();
    setStatus("Join the current game");
  }
}

async function placeFleet() {
  if (!state.playerId || state.draftShips.length !== fleet.length) {
    setStatus("Place the full fleet first");
    return;
  }
  try {
    state.view = await api("/api/place", {
      method: "POST",
      body: JSON.stringify({ playerId: state.playerId, ships: state.draftShips }),
    });
    render();
  } catch (error) {
    setStatus(error.message);
  }
}

async function shoot(row, col) {
  if (!state.view || state.view.phase !== "playing" || state.view.turn !== state.view.playerNumber) {
    return;
  }
  try {
    const data = await api("/api/shoot", {
      method: "POST",
      body: JSON.stringify({ playerId: state.playerId, row, col }),
    });
    state.view = data.view;
    render();
  } catch (error) {
    setStatus(error.message);
  }
}

async function resetGame() {
  await api("/api/reset", { method: "POST" });
  localStorage.removeItem(storageKey);
  state.playerId = "";
  state.playerNumber = 0;
  state.view = null;
  state.draftShips = [];
  render();
}

async function api(url, options = {}) {
  const response = await fetch(url, {
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  const data = await response.json();
  if (!response.ok) {
    throw new Error(data.error || "Request failed");
  }
  return data;
}

function render() {
  joinPanel.hidden = Boolean(state.playerId);
  autoPlaceButton.disabled = !state.playerId || Boolean(state.view && state.view.youReady);
  readyButton.disabled = !state.playerId || Boolean(state.view && state.view.youReady);
  playerBadge.textContent = state.playerNumber ? `Player ${state.playerNumber}` : "Not joined";

  if (!state.playerId) {
    setStatus("Join from two browser windows to start");
    turnBadge.textContent = "Waiting";
    renderEmptyBoards();
    return;
  }

  if (!state.view) {
    setStatus("Connecting...");
    renderDraftBoard();
    renderBoard(enemyBoardEl, emptyBoard(), () => {});
    return;
  }

  setStatus(state.view.message);
  turnBadge.textContent = state.view.phase === "playing" ? `Turn: player ${state.view.turn}` : state.view.phase;
  if (state.view.youReady) {
    renderBoard(ownBoardEl, state.view.ownBoard, () => {});
  } else {
    renderDraftBoard();
  }
  renderBoard(enemyBoardEl, state.view.enemyBoard, shoot);
}

function renderEmptyBoards() {
  renderBoard(ownBoardEl, emptyBoard(), () => {});
  renderBoard(enemyBoardEl, emptyBoard(), () => {});
}

function renderDraftBoard() {
  const board = emptyBoard();
  for (const ship of state.draftShips) {
    for (const cell of ship.cells) {
      board[cell.row][cell.col].state = "placed";
    }
  }
  renderBoard(ownBoardEl, board, () => {});
}

function renderBoard(container, board, onClick) {
  container.replaceChildren();
  for (let row = 0; row < boardSize; row++) {
    for (let col = 0; col < boardSize; col++) {
      const cell = document.createElement("button");
      cell.type = "button";
      cell.className = `cell ${board[row][col].state}`;
      cell.ariaLabel = `${row + 1}:${col + 1}`;
      cell.disabled = container === ownBoardEl || board[row][col].state !== "empty";
      cell.addEventListener("click", () => onClick(row, col));
      container.append(cell);
    }
  }
}

function emptyBoard() {
  return Array.from({ length: boardSize }, () =>
    Array.from({ length: boardSize }, () => ({ state: "empty" })),
  );
}

function generateFleet() {
  for (let attempt = 0; attempt < 200; attempt++) {
    const occupied = new Set();
    const ships = [];
    let failed = false;

    for (const size of fleet) {
      const ship = placeRandomShip(size, occupied);
      if (!ship) {
        failed = true;
        break;
      }
      ships.push({ cells: ship });
      for (const cell of ship) {
        occupied.add(key(cell.row, cell.col));
      }
    }

    if (!failed) {
      return ships;
    }
  }
  return [];
}

function placeRandomShip(size, occupied) {
  for (let attempt = 0; attempt < 100; attempt++) {
    const horizontal = Math.random() > 0.5;
    const row = randomInt(horizontal ? boardSize : boardSize - size + 1);
    const col = randomInt(horizontal ? boardSize - size + 1 : boardSize);
    const cells = Array.from({ length: size }, (_, index) => ({
      row: row + (horizontal ? 0 : index),
      col: col + (horizontal ? index : 0),
    }));
    if (cells.every((cell) => canUseCell(cell, occupied))) {
      return cells;
    }
  }
  return null;
}

function canUseCell(cell, occupied) {
  for (let row = cell.row - 1; row <= cell.row + 1; row++) {
    for (let col = cell.col - 1; col <= cell.col + 1; col++) {
      if (occupied.has(key(row, col))) {
        return false;
      }
    }
  }
  return true;
}

function randomInt(max) {
  return Math.floor(Math.random() * max);
}

function key(row, col) {
  return `${row}:${col}`;
}

function setStatus(message) {
  statusEl.textContent = message;
}
