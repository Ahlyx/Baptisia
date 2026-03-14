package main

import (
	"baptisia/physics"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	sensorFile   = filepath.Join(os.TempDir(), "baptisia_sensors.txt")
	actuatorFile = filepath.Join(os.TempDir(), "baptisia_actuator.txt")
	controlBin   = filepath.Join(".", "hal", "pump_control")
)

func writeSensors(e *physics.Engine) error {
	psi, flow, temp := e.Pump.Readings()
	content := fmt.Sprintf("%d %d %d\n", psi, flow, temp)
	return os.WriteFile(sensorFile, []byte(content), 0644)
}

func readActuator() (physics.PumpState, error) {
	data, err := os.ReadFile(actuatorFile)
	if err != nil {
		return physics.StateIdle, fmt.Errorf("reading actuator file: %w", err)
	}
	var state int
	if _, err := fmt.Sscanf(string(data), "%d", &state); err != nil {
		return physics.StateIdle, fmt.Errorf("parsing actuator value: %w", err)
	}
	return physics.PumpState(state), nil
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

		cmd := exec.Command(controlBin)
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error executing control loop binary: %v\n", err)
			return
		}

		state, err := readActuator()
		if err != nil {
			fmt.Println("Error reading actuator:", err)
			return
		}
		engine.Pump.SetState(state)

		psi, flow, temp := engine.Pump.Readings()
		fmt.Printf("[CYCLE %d] psi=%d flow=%d temp=%d state=%s\n",
			engine.Cycles, psi, flow, temp, engine.Pump.Status())

		engine.Step()
		time.Sleep(time.Duration(engine.CycleMS) * time.Millisecond)
	}

	engine.PrintSummary()
}
