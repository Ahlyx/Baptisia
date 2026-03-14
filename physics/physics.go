package physics

// PumpState mirrors the enum from generated C
type PumpState int

const (
	StateIdle        PumpState = 0
	StateRunning     PumpState = 1
	StateFault       PumpState = 2
	StateMaintenance PumpState = 3
)

// PumpConfig holds the physical constants for this pump
type PumpConfig struct {
	MaxPSI         float64 // safety cutoff
	MaxFlow        float64 // safety cutoff
	MaxTemp        float64 // safety cutoff
	PipeResistance float64 // how fast pressure builds (higher = faster)
	CoolRate       float64 // how fast temp drops when idle
	HeatRate       float64 // how fast temp rises when running
	PressureBleed  float64 // how fast pressure drops when idle
	FlowRampUp     float64 // how fast flow increases when running
	FlowRampDown   float64 // how fast flow drops when idle
}

// DefaultPumpConfig returns a realistic pump configuration
func DefaultPumpConfig() PumpConfig {
    return PumpConfig{
        MaxPSI:         100.0,
        MaxFlow:        500.0,  // match sensor range
        MaxTemp:        90.0,
        PipeResistance: 0.8,
        CoolRate:       0.5,
        HeatRate:       0.3,
        PressureBleed:  3.0,
        FlowRampUp:     8.0,   // slower ramp so PSI builds first
        FlowRampDown:   10.0,
    }
}

// PumpPhysics holds the current physical state of the pump
type PumpPhysics struct {
	Config    PumpConfig
	PSI       float64 // current pressure
	Flow      float64 // current flow rate
	Temp      float64 // current temperature
	RunCycles int     // how many cycles pump has been running
	State     PumpState
	Latched	  bool
}

// NewPump creates a new pump simulation starting at rest
func NewPump(config PumpConfig) *PumpPhysics {
	return &PumpPhysics{
		Config: config,
		PSI:    0.0,
		Flow:   0.0,
		Temp:   20.0, // ambient temperature
		State:  StateIdle,
	}
}

// SetState updates the pump state from the control loop
func (p *PumpPhysics) SetState(state PumpState) {
    // once faulted, ignore any attempt to return to RUNNING or IDLE
    if p.Latched && state != StateFault {
        return
    }
    if state == StateFault {
        p.Latched = true
    }
    p.State = state
}

// Step advances the physics simulation by one cycle
func (p *PumpPhysics) Step() {
	switch p.State {
	case StateRunning:
		p.RunCycles++

		// flow ramps up toward max
		p.Flow += p.Config.FlowRampUp
		if p.Flow > p.Config.MaxFlow*0.8 {
			p.Flow = p.Config.MaxFlow * 0.8
		}

		// pressure builds based on flow and pipe resistance
		p.PSI += p.Flow * p.Config.PipeResistance * 0.01
		if p.PSI > p.Config.MaxPSI*1.2 {
			p.PSI = p.Config.MaxPSI * 1.2 // cap at 20% over max
		}

		// temperature rises with run time
		p.Temp += p.Config.HeatRate
		if p.Temp > p.Config.MaxTemp*1.1 {
			p.Temp = p.Config.MaxTemp * 1.1
		}

	case StateIdle:
		p.RunCycles = 0

		// flow drops
		p.Flow -= p.Config.FlowRampDown
		if p.Flow < 0 {
			p.Flow = 0
		}

		// pressure bleeds off
		p.PSI -= p.Config.PressureBleed
		if p.PSI < 0 {
			p.PSI = 0
		}

		// temperature cools toward ambient
		if p.Temp > 20.0 {
			p.Temp -= p.Config.CoolRate
		}

	case StateFault:
		p.RunCycles = 0

		// immediate stop — flow and pressure drop fast
		p.Flow -= p.Config.FlowRampDown * 2
		if p.Flow < 0 {
			p.Flow = 0
		}

		p.PSI -= p.Config.PressureBleed * 2
		if p.PSI < 0 {
			p.PSI = 0
		}

		// temp keeps rising briefly after fault (thermal inertia)
		if p.Temp > 20.0 {
			p.Temp -= p.Config.CoolRate * 0.5
		}
	}
}

// Readings returns integer sensor readings for the HAL
func (p *PumpPhysics) Readings() (psi int32, flow int32, temp int32) {
	return int32(p.PSI), int32(p.Flow), int32(p.Temp)
}

// Status prints the current physical state for simulation output
func (p *PumpPhysics) Status() string {
	stateName := map[PumpState]string{
		StateIdle:        "IDLE",
		StateRunning:     "RUNNING",
		StateFault:       "FAULT",
		StateMaintenance: "MAINTENANCE",
	}
	return stateName[p.State]
}
