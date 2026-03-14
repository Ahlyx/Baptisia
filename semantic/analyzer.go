package semantic

import (
	"baptisia/ast"
	"fmt"
	"strconv"
)

// Check performs semantic validation on a parsed device.
// It returns a slice of error strings; an empty slice means valid.
func Check(device *ast.DeviceNode) []string {
	var errors []string

	if device == nil {
		return append(errors, "parsed program is missing a device declaration")
	}

	if device.Boot == nil {
		errors = append(errors, fmt.Sprintf("device %q is missing required block: boot", device.Name))
	}
	if device.Inputs == nil {
		errors = append(errors, fmt.Sprintf("device %q is missing required block: inputs", device.Name))
	}
	if device.Outputs == nil {
		errors = append(errors, fmt.Sprintf("device %q is missing required block: outputs", device.Name))
	}
	if device.Safety == nil {
		errors = append(errors, fmt.Sprintf("device %q is missing required block: safety", device.Name))
	}
	if device.Failsafe == nil {
		errors = append(errors, fmt.Sprintf("device %q is missing required block: failsafe", device.Name))
	}
	if device.Control == nil {
		errors = append(errors, fmt.Sprintf("device %q is missing required block: control", device.Name))
	}

	known := map[string]struct{}{
		"true":  {},
		"false": {},
		"off":   {},
		"on":    {},
	}

	if device.Vars != nil {
		for _, v := range device.Vars.Vars {
			known[v.Name] = struct{}{}
		}
	}
	if device.Consts != nil {
		for _, c := range device.Consts.Constants {
			known[c.Name] = struct{}{}
		}
	}
	if device.States != nil {
		for _, s := range device.States.Names {
			known[s] = struct{}{}
		}
	}
	if device.Outputs != nil {
		for _, o := range device.Outputs.Outputs {
			known[o.Name] = struct{}{}
		}
	}

	if device.Boot != nil {
		errors = append(errors, checkStatements(device.Name, "boot", device.Boot.Statements, known)...)
	}
	if device.Safety != nil {
		errors = append(errors, checkStatements(device.Name, "safety", device.Safety.Statements, known)...)
	}
	if device.Control != nil {
		errors = append(errors, checkStatements(device.Name, "control", device.Control.Statements, known)...)
	}

	return errors
}

func checkStatements(deviceName, block string, stmts []ast.Node, known map[string]struct{}) []string {
	var errors []string

	for _, node := range stmts {
		switch stmt := node.(type) {
		case *ast.AssignStatement:
			if !isKnownName(stmt.Name, known) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.Name))
			}
			if !isKnownName(stmt.Value, known) && !isLiteral(stmt.Value) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.Value))
			}
		case *ast.OutputCall:
			if !isKnownName(stmt.Relay, known) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.Relay))
			}
			if !isKnownName(stmt.State, known) && !isLiteral(stmt.State) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.State))
			}
		case *ast.IfStatement:
			if !isKnownName(stmt.Left, known) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.Left))
			}
			if !isKnownName(stmt.Right, known) && !isLiteral(stmt.Right) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.Right))
			}
			errors = append(errors, checkStatements(deviceName, block, []ast.Node{stmt.Then}, known)...)
		case *ast.IfOrStatement:
			if !isKnownName(stmt.LeftVar, known) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.LeftVar))
			}
			if !isKnownName(stmt.LeftVal, known) && !isLiteral(stmt.LeftVal) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.LeftVal))
			}
			if !isKnownName(stmt.RightVar, known) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.RightVar))
			}
			if !isKnownName(stmt.RightVal, known) && !isLiteral(stmt.RightVal) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.RightVal))
			}
			errors = append(errors, checkStatements(deviceName, block, []ast.Node{stmt.Then}, known)...)
		case *ast.IfElseStatement:
			if !isKnownName(stmt.LeftVar, known) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.LeftVar))
			}
			if !isKnownName(stmt.LeftVal, known) && !isLiteral(stmt.LeftVal) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.LeftVal))
			}
			if !isKnownName(stmt.RightVar, known) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.RightVar))
			}
			if !isKnownName(stmt.RightVal, known) && !isLiteral(stmt.RightVal) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.RightVal))
			}
			errors = append(errors, checkStatements(deviceName, block, []ast.Node{stmt.Then, stmt.Else}, known)...)
		case *ast.IfOrElseStatement:
			if !isKnownName(stmt.LeftVar, known) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.LeftVar))
			}
			if !isKnownName(stmt.LeftVal, known) && !isLiteral(stmt.LeftVal) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.LeftVal))
			}
			if !isKnownName(stmt.RightVar, known) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.RightVar))
			}
			if !isKnownName(stmt.RightVal, known) && !isLiteral(stmt.RightVal) {
				errors = append(errors, fmt.Sprintf("device %q has undefined identifier in %s: %q", deviceName, block, stmt.RightVal))
			}
			errors = append(errors, checkStatements(deviceName, block, []ast.Node{stmt.Then, stmt.Else}, known)...)
		}
	}

	return errors
}

func isKnownName(name string, known map[string]struct{}) bool {
	_, ok := known[name]
	return ok
}

func isLiteral(v string) bool {
	if _, err := strconv.Atoi(v); err == nil {
		return true
	}
	if _, err := strconv.ParseFloat(v, 64); err == nil {
		return true
	}
	if len(v) > 2 && v[len(v)-2:] == "ms" {
		if _, err := strconv.Atoi(v[:len(v)-2]); err == nil {
			return true
		}
	}
	return false
}
