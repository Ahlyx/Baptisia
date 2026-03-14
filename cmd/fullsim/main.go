package main

import (
	"baptisia/physics"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	sensorFile   = "/tmp/baptisia_sensors.txt"
	actuatorFile = "/tmp/baptisia_actuator.txt"
	controlBin   = "/home/ahlyx/Baptisia/hal/pump_control"
)

func writeSensors(e *physics.Engine) error {
	psi, flow, temp := e.Pump.Readings()
	content := fmt.Sprintf("%d %d %d\n", psi, flow, temp)
	return os.WriteFile(sensorFile, []byte(content), 0644)
}

func readActuator() physics.PumpState {
	data, err := os.ReadFile(actuatorFile)
	if err != nil {
		return physics.StateIdle
	}
	var state int
	fmt.Sscanf(string(data), "%d", &state)
	return physics.PumpState(state)
}

func main() {
	fmt.Println("=== Baptisia Full Simulation ===")
	fmt.Println("Physics engine + compiled Baptisia control loop")
	fmt.Println()

	engine := physics.NewEngine(80, 100, true)

	for engine.Cycles < engine.MaxCycles {
		if err := writeSensors(engine); err != nil {
			fmt.Println("Error writing sensors:", err)
			return
		}

		// run one cycle of the compiled Baptisia control loop
		cmd := exec.Command(controlBin)
		cmd.Stderr = os.Stderr
		cmd.Run()

		// read actuator decision and update physics engine
		state := readActuator()
		engine.Pump.SetState(state)

		// print engine state — single source of truth
		psi, flow, temp := engine.Pump.Readings()
		fmt.Printf("[CYCLE %d] psi=%d flow=%d temp=%d state=%s\n",
			engine.Cycles, psi, flow, temp, engine.Pump.Status())

		engine.Step()
		time.Sleep(time.Duration(engine.CycleMS) * time.Millisecond)
	}

	engine.PrintSummary()
}