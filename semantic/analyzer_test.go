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
