package codegen

import (
	"baptisia/ast"
	"fmt"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name     string
		device   *ast.DeviceNode
		emitMain bool
		want     string
	}{
		{
			name: "minimal device without states emitMain false",
			device: &ast.DeviceNode{
				Name:     "motor",
				Watchdog: "500ms",
				Cycle:    "100ms",
				Consts: &ast.ConstNode{Constants: []ast.ConstDecl{{Name: "max_speed", Value: "500"}, {Name: "max_temp", Value: "100"}}},
				Vars: &ast.VarsNode{Vars: []ast.VarDecl{
					{Volatile: true, TypeName: "BOOL", Name: "state", Value: "false"},
					{Volatile: true, TypeName: "I32", Name: "speed", Value: "0"},
					{Volatile: true, TypeName: "F32", Name: "temp", Value: "0.0"},
				}},
				Boot: &ast.BootNode{Statements: []ast.Node{
					&ast.AssignStatement{Name: "state", Value: "false"},
					&ast.OutputCall{Relay: "motor_relay", State: "off"},
				}},
				Inputs: &ast.InputsNode{Inputs: []ast.SensorAssign{{Name: "speed", Pin: "A0", Min: "0", Max: "1000"}, {Name: "temp", Pin: "A1", Min: "-40", Max: "150"}}},
				Outputs: &ast.OutputsNode{Outputs: []ast.ActuatorDecl{
					{Name: "motor_relay", Pin: "D0"},
				}},
				Safety: &ast.SafetyNode{Statements: []ast.Node{
					&ast.IfStatement{Left: "speed", Operator: ">=", Right: "max_speed", Then: &ast.AssignStatement{Name: "state", Value: "false"}},
				}},
				Failsafe: &ast.FailsafeNode{Statements: []ast.Node{
					&ast.AssignStatement{Name: "state", Value: "false"},
					&ast.OutputCall{Relay: "motor_relay", State: "off"},
				}},
				Control: &ast.ControlNode{Statements: []ast.Node{
					&ast.IfElseStatement{LeftVar: "speed", LeftOp: "<", LeftVal: "max_speed", RightVar: "temp", RightOp: "<", RightVal: "max_temp", Then: &ast.AssignStatement{Name: "state", Value: "true"}, Else: &ast.AssignStatement{Name: "state", Value: "false"}},
				}},
			},
			emitMain: false,
			want: `#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include "hal.h"

#define WATCHDOG_MS 500
#define CYCLE_MS 100

// Constants
#define MAX_SPEED 500
#define MAX_TEMP 100

// Variable declarations
volatile bool state = false;
volatile int32_t speed = 0;
volatile float temp = 0.0;

// Sensor validation ranges
#define SPEED_MIN 0
#define SPEED_MAX 1000
#define TEMP_MIN -40
#define TEMP_MAX 150

// Boot initialization - runs once on startup
void motor_boot(void) {
    state = false;
    write_actuator_motor_relay(0);
}

// Failsafe - called when safety conditions are violated
void motor_failsafe(void) {
    state = false;
    write_actuator_motor_relay(0);
}

// Main control loop for motor
void motor_loop(void) {
    watchdog_reset();

    // Read and validate inputs
    speed = read_sensor_A0();
    if (speed < SPEED_MIN || speed > SPEED_MAX) { motor_failsafe(); return; }
    temp = read_sensor_A1();
    if (temp < TEMP_MIN || temp > TEMP_MAX) { motor_failsafe(); return; }

    // Safety checks
    if (speed >= MAX_SPEED) {
        motor_failsafe();
        return;
    }

    // Control logic
    if (speed < MAX_SPEED && temp < MAX_TEMP) {
        state = true;
    } else {
        state = false;
    }

    // Write outputs
    write_actuator_D0(state);
}

`,
		},
		{
			name: "device with states and OR safety emitMain false",
			device: &ast.DeviceNode{
				Name:     "pump",
				Watchdog: "500ms",
				Cycle:    "100ms",
				Consts: &ast.ConstNode{Constants: []ast.ConstDecl{{Name: "max_psi", Value: "100"}, {Name: "max_flow", Value: "500"}}},
				States: &ast.StatesNode{Names: []string{"IDLE", "RUNNING", "FAULT", "MAINTENANCE"}},
				Vars: &ast.VarsNode{Vars: []ast.VarDecl{
					{Volatile: true, TypeName: "I32", Name: "psi", Value: "0"},
					{Volatile: true, TypeName: "I32", Name: "flow", Value: "0"},
				}},
				Boot: &ast.BootNode{Statements: []ast.Node{
					&ast.AssignStatement{Name: "state", Value: "IDLE"},
					&ast.OutputCall{Relay: "motor_relay", State: "off"},
				}},
				Inputs: &ast.InputsNode{Inputs: []ast.SensorAssign{{Name: "psi", Pin: "A0", Min: "0", Max: "500"}, {Name: "flow", Pin: "A1", Min: "0", Max: "500"}}},
				Outputs: &ast.OutputsNode{Outputs: []ast.ActuatorDecl{{Name: "motor_relay", Pin: "D0"}}},
				Safety: &ast.SafetyNode{Statements: []ast.Node{
					&ast.IfOrStatement{LeftVar: "psi", LeftOp: ">=", LeftVal: "max_psi", RightVar: "flow", RightOp: ">=", RightVal: "max_flow", Then: &ast.AssignStatement{Name: "state", Value: "FAULT"}},
				}},
				Failsafe: &ast.FailsafeNode{Statements: []ast.Node{
					&ast.AssignStatement{Name: "state", Value: "FAULT"},
					&ast.OutputCall{Relay: "motor_relay", State: "off"},
				}},
				Control: &ast.ControlNode{Statements: []ast.Node{
					&ast.IfElseStatement{LeftVar: "psi", LeftOp: "<", LeftVal: "max_psi", RightVar: "flow", RightOp: "<", RightVal: "max_flow", Then: &ast.AssignStatement{Name: "state", Value: "RUNNING"}, Else: &ast.AssignStatement{Name: "state", Value: "IDLE"}},
				}},
			},
			emitMain: false,
			want: `#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include "hal.h"

#define WATCHDOG_MS 500
#define CYCLE_MS 100

// Constants
#define MAX_PSI 100
#define MAX_FLOW 500

typedef enum {
    PUMP_IDLE,
    PUMP_RUNNING,
    PUMP_FAULT,
    PUMP_MAINTENANCE
} pump_state_t;

volatile pump_state_t state = PUMP_IDLE;

// Variable declarations
volatile int32_t psi = 0;
volatile int32_t flow = 0;

// Sensor validation ranges
#define PSI_MIN 0
#define PSI_MAX 500
#define FLOW_MIN 0
#define FLOW_MAX 500

// Boot initialization - runs once on startup
void pump_boot(void) {
    state = PUMP_IDLE;
    write_actuator_motor_relay(0);
}

// Failsafe - called when safety conditions are violated
void pump_failsafe(void) {
    state = PUMP_FAULT;
    write_actuator_motor_relay(0);
}

// Main control loop for pump
void pump_loop(void) {
    watchdog_reset();

    // Read and validate inputs
    psi = read_sensor_A0();
    if (psi < PSI_MIN || psi > PSI_MAX) { pump_failsafe(); return; }
    flow = read_sensor_A1();
    if (flow < FLOW_MIN || flow > FLOW_MAX) { pump_failsafe(); return; }

    // Safety checks
    if (psi >= MAX_PSI || flow >= MAX_FLOW) {
        pump_failsafe();
        return;
    }

    // Control logic
    if (psi < MAX_PSI && flow < MAX_FLOW) {
        state = PUMP_RUNNING;
    } else {
        state = PUMP_IDLE;
    }

    // Write outputs
    write_actuator_D0(state);
}

`,
		},
		{
			name: "emitMain true includes main loop",
			device: &ast.DeviceNode{
				Name:     "mini",
				Watchdog: "100ms",
				Cycle:    "10ms",
				Vars: &ast.VarsNode{Vars: []ast.VarDecl{
					{Volatile: true, TypeName: "BOOL", Name: "state", Value: "false"},
				}},
				Boot: &ast.BootNode{Statements: []ast.Node{
					&ast.AssignStatement{Name: "state", Value: "true"},
				}},
			},
			emitMain: true,
			want: `#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include "hal.h"

#define WATCHDOG_MS 100
#define CYCLE_MS 10

// Variable declarations
volatile bool state = false;

// Boot initialization - runs once on startup
void mini_boot(void) {
    state = true;
}

// Main control loop for mini
void mini_loop(void) {
    watchdog_reset();

}

int main(void) {
    mini_boot();
    while (1) {
        mini_loop();
        delay_ms(CYCLE_MS);
    }
    return 0;
}
`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := Generate(tc.device, tc.emitMain)
			if got != tc.want {
				t.Errorf("Generate() mismatch:\n%s\n\n--- got ---\n%s\n--- want ---\n%s", firstDiff(got, tc.want), got, tc.want)
			}
		})
	}
}

func firstDiff(got, want string) string {
	if got == want {
		return "no diff"
	}
	max := len(got)
	if len(want) < max {
		max = len(want)
	}
	for i := 0; i < max; i++ {
		if got[i] != want[i] {
			gStart := i - 20
			if gStart < 0 {
				gStart = 0
			}
			gEnd := i + 20
			if gEnd > len(got) {
				gEnd = len(got)
			}
			wStart := i - 20
			if wStart < 0 {
				wStart = 0
			}
			wEnd := i + 20
			if wEnd > len(want) {
				wEnd = len(want)
			}
			return fmt.Sprintf("first difference at byte %d\n got context: %q\nwant context: %q", i, got[gStart:gEnd], want[wStart:wEnd])
		}
	}
	return fmt.Sprintf("length differs: got=%d want=%d (prefix equal)", len(got), len(want))
}

func TestFirstDiffNoDiff(t *testing.T) {
	if diff := firstDiff("same", "same"); !strings.Contains(diff, "no diff") {
		t.Errorf("expected no diff, got %q", diff)
	}
}
