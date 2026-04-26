package model

import (
	"errors"
	"fmt"
	"sort"
)

const BoardSize = 10

var Fleet = []int{5, 4, 3, 3, 2}

type Coord struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

type ShipPlacement struct {
	Cells []Coord `json:"cells"`
}

type ShotResult struct {
	Hit    bool    `json:"hit"`
	Sunk   bool    `json:"sunk"`
	Winner *int    `json:"winner,omitempty"`
	Ship   []Coord `json:"ship,omitempty"`
}

type Game struct {
	Players [2]*Player
	Turn    int
	Winner  *int
}

type Player struct {
	ID    string
	Name  string
	Ready bool
	Ships []Ship
	Shots map[Coord]Shot
}

type Ship struct {
	Cells []Coord
	Hits  map[Coord]bool
}

type Shot struct {
	Hit  bool
	Sunk bool
}

type View struct {
	PlayerNumber int      `json:"playerNumber"`
	Phase        string   `json:"phase"`
	Turn         int      `json:"turn"`
	Winner       *int     `json:"winner,omitempty"`
	YouReady     bool     `json:"youReady"`
	EnemyReady   bool     `json:"enemyReady"`
	OwnBoard     [][]Cell `json:"ownBoard"`
	EnemyBoard   [][]Cell `json:"enemyBoard"`
	Message      string   `json:"message"`
}

type Cell struct {
	State string `json:"state"`
}

func NewGame() *Game {
	return &Game{Turn: 0}
}

func (g *Game) Join(id, name string) (int, error) {
	for i, player := range g.Players {
		if player != nil && player.ID == id {
			return i, nil
		}
	}
	for i := range g.Players {
		if g.Players[i] == nil {
			if name == "" {
				name = fmt.Sprintf("Player %d", i+1)
			}
			g.Players[i] = &Player{ID: id, Name: name, Shots: map[Coord]Shot{}}
			return i, nil
		}
	}
	return -1, errors.New("game already has two players")
}

func (g *Game) PlaceFleet(playerID string, placements []ShipPlacement) error {
	playerIndex, err := g.playerIndex(playerID)
	if err != nil {
		return err
	}
	if g.Winner != nil {
		return errors.New("game is already finished")
	}
	if len(placements) != len(Fleet) {
		return fmt.Errorf("expected %d ships", len(Fleet))
	}

	sortedPlacements := append([]ShipPlacement(nil), placements...)
	sort.Slice(sortedPlacements, func(i, j int) bool {
		return len(sortedPlacements[i].Cells) > len(sortedPlacements[j].Cells)
	})

	occupied := map[Coord]bool{}
	ships := make([]Ship, 0, len(Fleet))
	for i, placement := range sortedPlacements {
		expected := Fleet[i]
		if err := validateShip(placement.Cells, expected, occupied); err != nil {
			return err
		}
		cells := normalizeCells(placement.Cells)
		for _, cell := range cells {
			occupied[cell] = true
		}
		ships = append(ships, Ship{Cells: cells, Hits: map[Coord]bool{}})
	}

	g.Players[playerIndex].Ships = ships
	g.Players[playerIndex].Ready = true
	return nil
}

func (g *Game) Shoot(playerID string, target Coord) (ShotResult, error) {
	playerIndex, err := g.playerIndex(playerID)
	if err != nil {
		return ShotResult{}, err
	}
	if !inBounds(target) {
		return ShotResult{}, errors.New("shot is outside the board")
	}
	if g.Phase() != "playing" {
		return ShotResult{}, errors.New("both players must place ships first")
	}
	if g.Winner != nil {
		return ShotResult{}, errors.New("game is already finished")
	}
	if playerIndex != g.Turn {
		return ShotResult{}, errors.New("it is not your turn")
	}

	opponentIndex := 1 - playerIndex
	opponent := g.Players[opponentIndex]
	if _, exists := opponent.Shots[target]; exists {
		return ShotResult{}, errors.New("this cell was already targeted")
	}

	result := ShotResult{}
	for shipIndex := range opponent.Ships {
		ship := &opponent.Ships[shipIndex]
		if ship.hasCell(target) {
			ship.Hits[target] = true
			result.Hit = true
			result.Sunk = ship.isSunk()
			if result.Sunk {
				result.Ship = append([]Coord(nil), ship.Cells...)
			}
			break
		}
	}

	opponent.Shots[target] = Shot{Hit: result.Hit, Sunk: result.Sunk}
	if result.Hit {
		if allSunk(opponent.Ships) {
			winner := playerIndex
			g.Winner = &winner
			result.Winner = &winner
		}
		return result, nil
	}

	g.Turn = opponentIndex
	return result, nil
}

func (g *Game) View(playerID string) (View, error) {
	playerIndex, err := g.playerIndex(playerID)
	if err != nil {
		return View{}, err
	}
	opponentIndex := 1 - playerIndex
	player := g.Players[playerIndex]
	opponent := g.Players[opponentIndex]

	view := View{
		PlayerNumber: playerIndex + 1,
		Phase:        g.Phase(),
		Turn:         g.Turn + 1,
		Winner:       playerNumberPtr(g.Winner),
		YouReady:     player.Ready,
		OwnBoard:     ownBoard(player),
		EnemyBoard:   enemyBoard(opponent),
	}
	if opponent != nil {
		view.EnemyReady = opponent.Ready
	}
	view.Message = messageFor(view)
	return view, nil
}

func (g *Game) Phase() string {
	if g.Winner != nil {
		return "finished"
	}
	if g.Players[0] == nil || g.Players[1] == nil {
		return "waiting"
	}
	if !g.Players[0].Ready || !g.Players[1].Ready {
		return "placing"
	}
	return "playing"
}

func (g *Game) playerIndex(playerID string) (int, error) {
	for i, player := range g.Players {
		if player != nil && player.ID == playerID {
			return i, nil
		}
	}
	return -1, errors.New("unknown player")
}

func validateShip(cells []Coord, expected int, occupied map[Coord]bool) error {
	if len(cells) != expected {
		return fmt.Errorf("ship must have %d cells", expected)
	}
	seen := map[Coord]bool{}
	for _, cell := range cells {
		if !inBounds(cell) {
			return errors.New("ship is outside the board")
		}
		if seen[cell] {
			return errors.New("ship contains duplicate cells")
		}
		seen[cell] = true
		if occupied[cell] {
			return errors.New("ships cannot overlap")
		}
		for row := cell.Row - 1; row <= cell.Row+1; row++ {
			for col := cell.Col - 1; col <= cell.Col+1; col++ {
				if occupied[Coord{Row: row, Col: col}] {
					return errors.New("ships cannot touch each other")
				}
			}
		}
	}

	normalized := normalizeCells(cells)
	sameRow := true
	sameCol := true
	for _, cell := range normalized {
		sameRow = sameRow && cell.Row == normalized[0].Row
		sameCol = sameCol && cell.Col == normalized[0].Col
	}
	if !sameRow && !sameCol {
		return errors.New("ship must be straight")
	}
	for i := 1; i < len(normalized); i++ {
		prev := normalized[i-1]
		current := normalized[i]
		if sameRow && current.Col != prev.Col+1 {
			return errors.New("ship cells must be contiguous")
		}
		if sameCol && current.Row != prev.Row+1 {
			return errors.New("ship cells must be contiguous")
		}
	}
	return nil
}

func normalizeCells(cells []Coord) []Coord {
	normalized := append([]Coord(nil), cells...)
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Row == normalized[j].Row {
			return normalized[i].Col < normalized[j].Col
		}
		return normalized[i].Row < normalized[j].Row
	})
	return normalized
}

func inBounds(cell Coord) bool {
	return cell.Row >= 0 && cell.Row < BoardSize && cell.Col >= 0 && cell.Col < BoardSize
}

func (s Ship) hasCell(target Coord) bool {
	for _, cell := range s.Cells {
		if cell == target {
			return true
		}
	}
	return false
}

func (s Ship) isSunk() bool {
	for _, cell := range s.Cells {
		if !s.Hits[cell] {
			return false
		}
	}
	return true
}

func allSunk(ships []Ship) bool {
	for _, ship := range ships {
		if !ship.isSunk() {
			return false
		}
	}
	return len(ships) > 0
}

func ownBoard(player *Player) [][]Cell {
	board := emptyBoard()
	if player == nil {
		return board
	}
	for _, ship := range player.Ships {
		for _, cell := range ship.Cells {
			board[cell.Row][cell.Col].State = "ship"
			if ship.Hits[cell] {
				board[cell.Row][cell.Col].State = "hit"
			}
		}
	}
	for cell, shot := range player.Shots {
		if !shot.Hit {
			board[cell.Row][cell.Col].State = "miss"
		}
	}
	return board
}

func enemyBoard(opponent *Player) [][]Cell {
	board := emptyBoard()
	if opponent == nil {
		return board
	}
	for cell, shot := range opponent.Shots {
		if shot.Hit {
			board[cell.Row][cell.Col].State = "hit"
		} else {
			board[cell.Row][cell.Col].State = "miss"
		}
	}
	for _, ship := range opponent.Ships {
		if ship.isSunk() {
			for _, cell := range ship.Cells {
				board[cell.Row][cell.Col].State = "sunk"
			}
		}
	}
	return board
}

func emptyBoard() [][]Cell {
	board := make([][]Cell, BoardSize)
	for row := range board {
		board[row] = make([]Cell, BoardSize)
		for col := range board[row] {
			board[row][col] = Cell{State: "empty"}
		}
	}
	return board
}

func playerNumberPtr(index *int) *int {
	if index == nil {
		return nil
	}
	number := *index + 1
	return &number
}

func messageFor(view View) string {
	switch view.Phase {
	case "waiting":
		return "Waiting for the second player"
	case "placing":
		if !view.YouReady {
			return "Place your fleet"
		}
		return "Waiting for opponent fleet"
	case "playing":
		if view.Turn == view.PlayerNumber {
			return "Your turn"
		}
		return "Opponent turn"
	case "finished":
		if view.Winner != nil && *view.Winner == view.PlayerNumber {
			return "You won"
		}
		return "You lost"
	default:
		return ""
	}
}
