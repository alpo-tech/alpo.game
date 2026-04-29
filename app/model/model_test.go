package model

import "testing"

func TestPlaceFleetAndShoot(t *testing.T) {
	game := NewGame()
	if _, err := game.Join("p1", "one"); err != nil {
		t.Fatal(err)
	}
	if _, err := game.Join("p2", "two"); err != nil {
		t.Fatal(err)
	}

	if err := game.PlaceFleet("p1", testFleetTop()); err != nil {
		t.Fatal(err)
	}
	if err := game.PlaceFleet("p2", testFleetTop()); err != nil {
		t.Fatal(err)
	}

	result, err := game.Shoot("p1", Coord{Row: 0, Col: 0})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Hit {
		t.Fatalf("expected hit")
	}
	if game.Turn != 0 {
		t.Fatalf("hit should keep turn, got %d", game.Turn)
	}

	result, err = game.Shoot("p1", Coord{Row: 9, Col: 9})
	if err != nil {
		t.Fatal(err)
	}
	if result.Hit {
		t.Fatalf("expected miss")
	}
	if game.Turn != 1 {
		t.Fatalf("miss should pass turn, got %d", game.Turn)
	}
}

func TestRejectTouchingShips(t *testing.T) {
	game := NewGame()
	if _, err := game.Join("p1", "one"); err != nil {
		t.Fatal(err)
	}

	badFleet := []ShipPlacement{
		{Cells: []Coord{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}}},
		{Cells: []Coord{{1, 0}, {1, 1}, {1, 2}, {1, 3}}},
		{Cells: []Coord{{3, 0}, {3, 1}, {3, 2}}},
		{Cells: []Coord{{5, 0}, {5, 1}, {5, 2}}},
		{Cells: []Coord{{7, 0}, {7, 1}}},
	}
	if err := game.PlaceFleet("p1", badFleet); err == nil {
		t.Fatalf("expected touching ships to be rejected")
	}
}

func TestViewHidesOpponentShips(t *testing.T) {
	game := NewGame()
	if _, err := game.Join("p1", "one"); err != nil {
		t.Fatal(err)
	}
	if _, err := game.Join("p2", "two"); err != nil {
		t.Fatal(err)
	}
	if err := game.PlaceFleet("p1", testFleetTop()); err != nil {
		t.Fatal(err)
	}
	if err := game.PlaceFleet("p2", testFleetTop()); err != nil {
		t.Fatal(err)
	}

	view, err := game.View("p1")
	if err != nil {
		t.Fatal(err)
	}
	for row := range view.EnemyBoard {
		for col := range view.EnemyBoard[row] {
			if view.EnemyBoard[row][col].State == "ship" {
				t.Fatalf("enemy ship leaked at %d,%d", row, col)
			}
		}
	}
}

func TestPlaceFleetRejectsReadyPlayer(t *testing.T) {
	game := NewGame()
	if _, err := game.Join("p1", "one"); err != nil {
		t.Fatal(err)
	}
	if err := game.PlaceFleet("p1", testFleetTop()); err != nil {
		t.Fatal(err)
	}

	if err := game.PlaceFleet("p1", testFleetBottom()); err == nil {
		t.Fatalf("expected replacing a ready fleet to be rejected")
	}
}

func testFleetTop() []ShipPlacement {
	return []ShipPlacement{
		{Cells: []Coord{{0, 0}, {0, 1}, {0, 2}, {0, 3}, {0, 4}}},
		{Cells: []Coord{{2, 0}, {2, 1}, {2, 2}, {2, 3}}},
		{Cells: []Coord{{4, 0}, {4, 1}, {4, 2}}},
		{Cells: []Coord{{6, 0}, {6, 1}, {6, 2}}},
		{Cells: []Coord{{8, 0}, {8, 1}}},
	}
}

func testFleetBottom() []ShipPlacement {
	return []ShipPlacement{
		{Cells: []Coord{{9, 5}, {9, 6}, {9, 7}, {9, 8}, {9, 9}}},
		{Cells: []Coord{{7, 6}, {7, 7}, {7, 8}, {7, 9}}},
		{Cells: []Coord{{5, 7}, {5, 8}, {5, 9}}},
		{Cells: []Coord{{3, 7}, {3, 8}, {3, 9}}},
		{Cells: []Coord{{1, 8}, {1, 9}}},
	}
}
