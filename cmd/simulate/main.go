package main

import (
	"baptisia/physics"
	"fmt"
)

func main() {
	fmt.Println("Starting Baptisia pump simulation...")

	engine := physics.NewEngine(60, 100, true)

	// start pump running so physics actually builds up
	engine.Pump.SetState(physics.StateRunning)

	engine.RunStandalone()
}
