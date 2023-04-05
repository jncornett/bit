package main

import "github.com/jncornett/bit"

type GameState struct{}

func NewGameState() *GameState {
	return &GameState{}
}

func (gs *GameState) Update(bit.Frame, bit.RenderTarget) {}
