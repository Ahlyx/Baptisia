package semantic

import (
	"baptisia/ast"
	"strings"
	"testing"
)

func TestCheck(t *testing.T) {
	base := func() *ast.DeviceNode {
		return &ast.DeviceNode{
			Name:     "motor",
			Boot:     &ast.BootNode{},
			Inputs:   &ast.InputsNode{},
			Outputs:  &ast.OutputsNode{},
			Safety:   &ast.SafetyNode{},
			Failsafe: &ast.FailsafeNode{},
			Control:  &ast.ControlNode{},
		}
	}

	tests := []struct {
		name       string
		device     *ast.DeviceNode
		wantClean  bool
		wantErrors []string
	}{
		{
			name:      "all mandatory blocks present",
			device:    base(),
			wantClean: true,
		},
		{
			name:       "missing boot",
			device:     func() *ast.DeviceNode { d := base(); d.Boot = nil; return d }(),
			wantErrors: []string{"boot"},
		},
		{
			name:       "missing safety",
			device:     func() *ast.DeviceNode { d := base(); d.Safety = nil; return d }(),
			wantErrors: []string{"safety"},
		},
		{
			name:       "missing failsafe",
			device:     func() *ast.DeviceNode { d := base(); d.Failsafe = nil; return d }(),
			wantErrors: []string{"failsafe"},
		},
		{
			name:       "missing control",
			device:     func() *ast.DeviceNode { d := base(); d.Control = nil; return d }(),
			wantErrors: []string{"control"},
		},
		{
			name:       "missing inputs",
			device:     func() *ast.DeviceNode { d := base(); d.Inputs = nil; return d }(),
			wantErrors: []string{"inputs"},
		},
		{
			name:       "missing outputs",
			device:     func() *ast.DeviceNode { d := base(); d.Outputs = nil; return d }(),
			wantErrors: []string{"outputs"},
		},
		{
			name: "undeclared variable in safety",
			device: &ast.DeviceNode{
				Name:     "motor",
				Vars:     &ast.VarsNode{Vars: []ast.VarDecl{{Name: "state"}}},
				Consts:   &ast.ConstNode{Constants: []ast.ConstDecl{{Name: "max_temp"}}},
				Boot:     &ast.BootNode{},
				Inputs:   &ast.InputsNode{},
				Outputs:  &ast.OutputsNode{},
				Failsafe: &ast.FailsafeNode{},
				Control:  &ast.ControlNode{},
				Safety: &ast.SafetyNode{Statements: []ast.Node{
					&ast.IfStatement{Left: "unknown", Operator: ">=", Right: "max_temp", Then: &ast.AssignStatement{Name: "state", Value: "false"}},
				}},
			},
			wantErrors: []string{"unknown"},
		},
		{
			name: "undeclared variable in control",
			device: &ast.DeviceNode{
				Name:     "motor",
				Vars:     &ast.VarsNode{Vars: []ast.VarDecl{{Name: "state"}, {Name: "temp"}}},
				Consts:   &ast.ConstNode{Constants: []ast.ConstDecl{{Name: "max_temp"}}},
				Boot:     &ast.BootNode{},
				Inputs:   &ast.InputsNode{},
				Outputs:  &ast.OutputsNode{},
				Safety:   &ast.SafetyNode{},
				Failsafe: &ast.FailsafeNode{},
				Control: &ast.ControlNode{Statements: []ast.Node{
					&ast.IfElseStatement{LeftVar: "temp", LeftOp: "<", LeftVal: "max_temp", RightVar: "missing", RightOp: "<", RightVal: "max_temp", Then: &ast.AssignStatement{Name: "state", Value: "true"}, Else: &ast.AssignStatement{Name: "state", Value: "false"}},
				}},
			},
			wantErrors: []string{"missing"},
		},
		{
			name: "declared vars and consts only",
			device: &ast.DeviceNode{
				Name:     "motor",
				Vars:     &ast.VarsNode{Vars: []ast.VarDecl{{Name: "state"}, {Name: "speed"}, {Name: "temp"}}},
				Consts:   &ast.ConstNode{Constants: []ast.ConstDecl{{Name: "max_speed"}, {Name: "max_temp"}}},
				Boot:     &ast.BootNode{Statements: []ast.Node{&ast.AssignStatement{Name: "state", Value: "false"}}},
				Inputs:   &ast.InputsNode{},
				Outputs:  &ast.OutputsNode{},
				Safety:   &ast.SafetyNode{Statements: []ast.Node{&ast.IfStatement{Left: "speed", Operator: ">=", Right: "max_speed", Then: &ast.AssignStatement{Name: "state", Value: "false"}}}},
				Failsafe: &ast.FailsafeNode{},
				Control:  &ast.ControlNode{Statements: []ast.Node{&ast.IfElseStatement{LeftVar: "speed", LeftOp: "<", LeftVal: "max_speed", RightVar: "temp", RightOp: "<", RightVal: "max_temp", Then: &ast.AssignStatement{Name: "state", Value: "true"}, Else: &ast.AssignStatement{Name: "state", Value: "false"}}}},
			},
			wantClean: true,
		},
		{
			name: "declared state name assignment value",
			device: &ast.DeviceNode{
				Name:     "pump",
				Vars:     &ast.VarsNode{Vars: []ast.VarDecl{{Name: "state"}, {Name: "psi"}, {Name: "flow"}}},
				Consts:   &ast.ConstNode{Constants: []ast.ConstDecl{{Name: "max_psi"}, {Name: "max_flow"}}},
				States:   &ast.StatesNode{Names: []string{"IDLE", "RUNNING", "FAULT"}},
				Boot:     &ast.BootNode{Statements: []ast.Node{&ast.AssignStatement{Name: "state", Value: "IDLE"}}},
				Inputs:   &ast.InputsNode{},
				Outputs:  &ast.OutputsNode{},
				Safety:   &ast.SafetyNode{Statements: []ast.Node{&ast.IfOrStatement{LeftVar: "psi", LeftOp: ">=", LeftVal: "max_psi", RightVar: "flow", RightOp: ">=", RightVal: "max_flow", Then: &ast.AssignStatement{Name: "state", Value: "FAULT"}}}},
				Failsafe: &ast.FailsafeNode{},
				Control:  &ast.ControlNode{Statements: []ast.Node{&ast.IfElseStatement{LeftVar: "psi", LeftOp: "<", LeftVal: "max_psi", RightVar: "flow", RightOp: "<", RightVal: "max_flow", Then: &ast.AssignStatement{Name: "state", Value: "RUNNING"}, Else: &ast.AssignStatement{Name: "state", Value: "IDLE"}}}},
			},
			wantClean: true,
		},
		{
			name: "multiple missing blocks",
			device: &ast.DeviceNode{
				Name:    "motor",
				Inputs:  &ast.InputsNode{},
				Outputs: &ast.OutputsNode{},
			},
			wantErrors: []string{"boot", "safety", "failsafe", "control"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errs := Check(tc.device)
			if tc.wantClean {
				if len(errs) != 0 {
					t.Fatalf("expected no errors, got %v", errs)
				}
				return
			}

			if len(errs) == 0 {
				t.Fatalf("expected errors, got none")
			}
			for _, want := range tc.wantErrors {
				found := false
				for _, err := range errs {
					if strings.Contains(err, want) {
						found = true
						break
					}
				}
				if !found {
					t.Fatalf("expected an error mentioning %q, got %v", want, errs)
				}
			}
			if len(tc.wantErrors) != len(errs) {
				t.Fatalf("expected %d errors, got %d: %v", len(tc.wantErrors), len(errs), errs)
			}
		})
	}
}
