package physics

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

// SensorValues holds the current readings exposed to the HAL
// These are read by the C control loop via the HAL layer
type SensorValues struct {
	PSI  int32
	Flow int32
	Temp int32
}

// Engine runs the physics simulation and manages the control loop
type Engine struct {
	Pump      *PumpPhysics
	Sensors   SensorValues
	Cycles    int
	MaxCycles int
	CycleMS   int
	Verbose   bool
}

// NewEngine creates a simulation engine with default pump config
func NewEngine(maxCycles int, cycleMS int, verbose bool) *Engine {
	return &Engine{
		Pump:      NewPump(DefaultPumpConfig()),
		MaxCycles: maxCycles,
		CycleMS:   cycleMS,
		Verbose:   verbose,
	}
}

// NewEngineWithConfig creates a simulation engine with custom pump config
func NewEngineWithConfig(config PumpConfig, maxCycles int, cycleMS int, verbose bool) *Engine {
	return &Engine{
		Pump:      NewPump(config),
		MaxCycles: maxCycles,
		CycleMS:   cycleMS,
		Verbose:   verbose,
	}
}

// UpdateSensors reads current physics state into sensor values
func (e *Engine) UpdateSensors() {
	psi, flow, temp := e.Pump.Readings()
	e.Sensors.PSI = psi
	e.Sensors.Flow = flow
	e.Sensors.Temp = temp
}

// SetPumpState updates the pump state from an external signal
// This is called by the HAL after the C control loop writes an actuator
func (e *Engine) SetPumpState(state PumpState) {
	e.Pump.SetState(state)
}

// Step advances one simulation cycle
func (e *Engine) Step() {
	e.Pump.Step()
	e.UpdateSensors()
	e.Cycles++
}

// PrintCycleHeader prints the cycle status for simulation output
func (e *Engine) PrintCycleHeader() {
	psi, flow, temp := e.Pump.Readings()
	fmt.Printf("\n[CYCLE %d] psi=%d flow=%d temp=%d state=%s\n",
		e.Cycles, psi, flow, temp, e.Pump.Status())
}

// PrintSummary prints a summary at the end of the simulation
func (e *Engine) PrintSummary() {
    fmt.Printf("\n=== Simulation Complete ===\n")
    fmt.Printf("Total cycles: %d\n", e.Cycles)
    fmt.Printf("Final state:  %s\n", e.Pump.Status())
    if e.Pump.Latched {
        fmt.Printf("Fault latched: YES — manual reset required\n")
    }
    psi, flow, temp := e.Pump.Readings()
    fmt.Printf("Final PSI:    %d\n", psi)
    fmt.Printf("Final flow:   %d\n", flow)
    fmt.Printf("Final temp:   %d°C\n", temp)
}

// RunStandalone runs the simulation entirely in Go without the C HAL
// Useful for testing the physics model before wiring up the C control loop
func (e *Engine) RunStandalone() {
	fmt.Println("=== Baptisia Physics Engine - Standalone Mode ===")
	fmt.Printf("Config: MaxPSI=%.0f MaxFlow=%.0f MaxTemp=%.0f\n",
		e.Pump.Config.MaxPSI,
		e.Pump.Config.MaxFlow,
		e.Pump.Config.MaxTemp)

	for e.Cycles < e.MaxCycles {
		e.PrintCycleHeader()

		// simple standalone control logic for testing
		// mirrors what the generated C would do
		psi, flow, _ := e.Pump.Readings()

		if int32(e.Pump.Config.MaxPSI) <= psi || int32(e.Pump.Config.MaxFlow) <= flow {
			// safety triggered
			fmt.Println("[SAFETY] Limit exceeded — FAULT")
			e.Pump.SetState(StateFault)
		} else if e.Pump.State == StateFault {
			// stay in fault
		} else if psi < int32(e.Pump.Config.MaxPSI) && flow < int32(e.Pump.Config.MaxFlow) {
			e.Pump.SetState(StateRunning)
		} else {
			e.Pump.SetState(StateIdle)
		}

		e.Step()
		time.Sleep(time.Duration(e.CycleMS) * time.Millisecond)
	}

	e.PrintSummary()
}

// WriteSharedState writes sensor values to a temp file for the C HAL to read
// This is the bridge between the Go physics engine and the C control loop
func (e *Engine) WriteSharedState(path string) error {
	content := fmt.Sprintf("%d %d %d\n",
		e.Sensors.PSI,
		e.Sensors.Flow,
		e.Sensors.Temp)
	return os.WriteFile(path, []byte(content), 0644)
}

// ReadActuatorState reads the actuator output written by the C control loop
func (e *Engine) ReadActuatorState(path string) (PumpState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return StateIdle, err
	}
	var state int
	fmt.Sscanf(string(data), "%d", &state)
	return PumpState(state), nil
}

// RunWithCBinary runs the physics engine alongside a compiled C control loop
// The C binary reads sensor values from a shared file and writes actuator state back
func (e *Engine) RunWithCBinary(binaryPath string, sensorFile string, actuatorFile string) {
	fmt.Println("=== Baptisia Physics Engine ===")
	fmt.Printf("Control binary: %s\n", binaryPath)

	for e.Cycles < e.MaxCycles {
		// write current sensor values for C to read
		if err := e.WriteSharedState(sensorFile); err != nil {
			fmt.Println("Error writing sensor state:", err)
			return
		}

		// run one cycle of the C control loop
		cmd := exec.Command(binaryPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Println("Control loop error:", err)
		}

		// read what the C loop decided
		state, err := e.ReadActuatorState(actuatorFile)
		if err == nil {
			e.Pump.SetState(state)
		}

		e.PrintCycleHeader()
		e.Step()
		time.Sleep(time.Duration(e.CycleMS) * time.Millisecond)
	}

	e.PrintSummary()
}
