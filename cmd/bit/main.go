package main

import "github.com/jncornett/bit"

func main() {
	state := NewGameState()
	bit.NewApp(bit.Update(state.Update)).Main()
}
