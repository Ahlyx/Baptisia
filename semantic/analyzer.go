package semantic

import (
	"baptisia/ast"
	"fmt"
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

	return errors
}
