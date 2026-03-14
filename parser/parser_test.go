package parser

import (
	"baptisia/ast"
	"baptisia/lexer"
	"testing"
)

func TestParserScenarios(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantErr   bool
		assertion func(t *testing.T, p *Parser, program *ast.Program)
	}{
		{
			name: "valid minimal device fields",
			input: `device motor : PLC {
	watchdog: 500ms
	cycle: 100ms
	boot { state = false }
	inputs { speed = val(sensor(A0), min: 0, max: 1000) }
	outputs { motor_relay = actuator(D0) }
	safety { if speed >= max_speed : state = false }
	failsafe { state = false output(motor_relay, off) }
	control { if speed < max_speed AND speed < max_speed : state = true else : state = false }
}`,
			assertion: func(t *testing.T, p *Parser, program *ast.Program) {
				if program.Device == nil {
					t.Fatalf("expected device, got nil")
				}
				if program.Device.Name != "motor" || program.Device.Target != "PLC" {
					t.Fatalf("unexpected device identity: %#v", program.Device)
				}
				if program.Device.Watchdog != "500ms" || program.Device.Cycle != "100ms" {
					t.Fatalf("unexpected timing values: watchdog=%q cycle=%q", program.Device.Watchdog, program.Device.Cycle)
				}
			},
		},
		{
			name: "single safety if parses",
			input: `device motor : PLC {
	watchdog: 500ms
	cycle: 100ms
	boot { state = false }
	inputs { speed = val(sensor(A0), min: 0, max: 1000) }
	outputs { motor_relay = actuator(D0) }
	safety { if speed >= max_speed : state = false }
	failsafe { state = false output(motor_relay, off) }
	control { if speed < max_speed AND speed < max_speed : state = true else : state = false }
}`,
			assertion: func(t *testing.T, p *Parser, program *ast.Program) {
				if program.Device.Safety == nil || len(program.Device.Safety.Statements) != 1 {
					t.Fatalf("expected one safety statement")
				}
				stmt, ok := program.Device.Safety.Statements[0].(*ast.IfStatement)
				if !ok {
					t.Fatalf("expected *ast.IfStatement, got %T", program.Device.Safety.Statements[0])
				}
				if stmt.Left != "speed" || stmt.Operator != ">=" || stmt.Right != "max_speed" {
					t.Fatalf("unexpected if fields: %#v", stmt)
				}
			},
		},
		{
			name: "or safety parses",
			input: `device pump : SCADA {
	watchdog: 500ms
	cycle: 100ms
	boot { state = IDLE }
	inputs {
		psi = val(sensor(A0), min: 0, max: 500)
		flow = val(sensor(A1), min: 0, max: 500)
	}
	outputs { motor_relay = actuator(D0) }
	safety { if psi >= max_psi OR flow >= max_flow : state = FAULT }
	failsafe { state = FAULT output(motor_relay, off) }
	control { if psi < max_psi AND flow < max_flow : state = RUNNING else : state = IDLE }
}`,
			assertion: func(t *testing.T, p *Parser, program *ast.Program) {
				if program.Device.Safety == nil || len(program.Device.Safety.Statements) != 1 {
					t.Fatalf("expected one safety statement")
				}
				stmt, ok := program.Device.Safety.Statements[0].(*ast.IfOrStatement)
				if !ok {
					t.Fatalf("expected *ast.IfOrStatement, got %T", program.Device.Safety.Statements[0])
				}
				if stmt.LeftVar != "psi" || stmt.LeftOp != ">=" || stmt.LeftVal != "max_psi" ||
					stmt.RightVar != "flow" || stmt.RightOp != ">=" || stmt.RightVal != "max_flow" {
					t.Fatalf("unexpected or-if fields: %#v", stmt)
				}
			},
		},
		{
			name: "and control parses",
			input: `device motor : PLC {
	watchdog: 500ms
	cycle: 100ms
	boot { state = false }
	inputs { speed = val(sensor(A0), min: 0, max: 1000) }
	outputs { motor_relay = actuator(D0) }
	safety { if speed >= max_speed : state = false }
	failsafe { state = false output(motor_relay, off) }
	control { if speed < max_speed AND temp < max_temp : state = true else : state = false }
}`,
			assertion: func(t *testing.T, p *Parser, program *ast.Program) {
				if program.Device.Control == nil || len(program.Device.Control.Statements) != 1 {
					t.Fatalf("expected one control statement")
				}
				stmt, ok := program.Device.Control.Statements[0].(*ast.IfElseStatement)
				if !ok {
					t.Fatalf("expected *ast.IfElseStatement, got %T", program.Device.Control.Statements[0])
				}
				if stmt.LeftVar != "speed" || stmt.LeftOp != "<" || stmt.LeftVal != "max_speed" ||
					stmt.RightVar != "temp" || stmt.RightOp != "<" || stmt.RightVal != "max_temp" {
					t.Fatalf("unexpected and-if fields: %#v", stmt)
				}
			},
		},
		{
			name: "unknown keyword in control produces parser error",
			input: `device motor : PLC {
	watchdog: 500ms
	cycle: 100ms
	boot { state = false }
	inputs { speed = val(sensor(A0), min: 0, max: 1000) }
	outputs { motor_relay = actuator(D0) }
	safety { if speed >= max_speed : state = false }
	failsafe { state = false output(motor_relay, off) }
	control { garbage 123 }
}`,
			wantErr: true,
		},
		{
			name: "missing closing brace produces parse error",
			input: `device motor : PLC {
	watchdog: 500ms
	cycle: 100ms
	boot { state = false }
	inputs { speed = val(sensor(A0), min: 0, max: 1000) }
	outputs { motor_relay = actuator(D0) }
	safety { if speed >= max_speed : state = false }
	failsafe { state = false output(motor_relay, off) }
	control { if speed < max_speed AND speed < max_speed : state = true else : state = false }
`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			l := lexer.New(tc.input)
			p := New(l)
			program := p.ParseProgram()

			if tc.wantErr {
				if len(p.Errors) == 0 {
					t.Fatalf("expected parser errors, got none")
				}
				return
			}

			if len(p.Errors) > 0 {
				t.Fatalf("unexpected parser errors: %v", p.Errors)
			}
			if tc.assertion != nil {
				tc.assertion(t, p, program)
			}
		})
	}
}
