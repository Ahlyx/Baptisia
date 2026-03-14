package codegen

import (
	"baptisia/ast"
	"fmt"
	"strings"
)

var typeMap = map[string]string{
	"BOOL": "bool",
	"I32":  "int32_t",
	"F32":  "float",
}

func Generate(device *ast.DeviceNode, emitMain bool) string {
	var out strings.Builder

	out.WriteString("#include <stdbool.h>\n")
	out.WriteString("#include <stdint.h>\n")
	out.WriteString("#include <stdio.h>\n")
	out.WriteString("#include \"hal.h\"\n\n")

	fmt.Fprintf(&out, "#define WATCHDOG_MS %s\n", strings.TrimSuffix(device.Watchdog, "ms"))
	fmt.Fprintf(&out, "#define CYCLE_MS %s\n\n", strings.TrimSuffix(device.Cycle, "ms"))

	if device.Consts != nil {
		out.WriteString("// Constants\n")
		for _, c := range device.Consts.Constants {
			fmt.Fprintf(&out, "#define %s %s\n", strings.ToUpper(c.Name), c.Value)
		}
		out.WriteString("\n")
	}

	if device.States != nil {
		upper := strings.ToUpper(device.Name)
		fmt.Fprintf(&out, "typedef enum {\n")
		for i, s := range device.States.Names {
			if i == len(device.States.Names)-1 {
				fmt.Fprintf(&out, "    %s_%s\n", upper, s)
			} else {
				fmt.Fprintf(&out, "    %s_%s,\n", upper, s)
			}
		}
		fmt.Fprintf(&out, "} %s_state_t;\n\n", device.Name)
		fmt.Fprintf(&out, "volatile %s_state_t state = %s_%s;\n\n",
			device.Name, upper, device.States.Names[0])
	}

	if device.Vars != nil {
		out.WriteString("// Variable declarations\n")
		for _, v := range device.Vars.Vars {
			cType := typeMap[v.TypeName]
			if v.Volatile {
				fmt.Fprintf(&out, "volatile %s %s = %s;\n", cType, v.Name, v.Value)
			} else {
				fmt.Fprintf(&out, "%s %s = %s;\n", cType, v.Name, v.Value)
			}
		}
		out.WriteString("\n")
	}

	if device.Inputs != nil {
		out.WriteString("// Sensor validation ranges\n")
		for _, s := range device.Inputs.Inputs {
			upper := strings.ToUpper(s.Name)
			fmt.Fprintf(&out, "#define %s_MIN %s\n", upper, s.Min)
			fmt.Fprintf(&out, "#define %s_MAX %s\n", upper, s.Max)
		}
		out.WriteString("\n")
	}

	if device.Boot != nil {
		out.WriteString("// Boot initialization - runs once on startup\n")
		fmt.Fprintf(&out, "void %s_boot(void) {\n", device.Name)
		for _, s := range device.Boot.Statements {
			out.WriteString(generateStatement(s, "    ", device.Name))
		}
		out.WriteString("}\n\n")
	}

	if device.Failsafe != nil {
		out.WriteString("// Failsafe - called when safety conditions are violated\n")
		fmt.Fprintf(&out, "void %s_failsafe(void) {\n", device.Name)
		for _, s := range device.Failsafe.Statements {
			out.WriteString(generateStatement(s, "    ", device.Name))
		}
		out.WriteString("}\n\n")
	}

	fmt.Fprintf(&out, "// Main control loop for %s\n", device.Name)
	fmt.Fprintf(&out, "void %s_loop(void) {\n", device.Name)
	out.WriteString("    watchdog_reset();\n\n")

	if device.Inputs != nil {
		out.WriteString("    // Read and validate inputs\n")
		for _, s := range device.Inputs.Inputs {
			upper := strings.ToUpper(s.Name)
			fmt.Fprintf(&out, "    %s = read_sensor_%s();\n", s.Name, s.Pin)
			fmt.Fprintf(&out, "    if (%s < %s_MIN || %s > %s_MAX) { %s_failsafe(); return; }\n",
				s.Name, upper, s.Name, upper, device.Name)
		}
		out.WriteString("\n")
	}

	if device.Safety != nil {
		out.WriteString("    // Safety checks\n")
		for _, s := range device.Safety.Statements {
			out.WriteString(generateStatement(s, "    ", device.Name))
		}
		out.WriteString("\n")
	}

	if device.Control != nil {
		out.WriteString("    // Control logic\n")
		for _, s := range device.Control.Statements {
			out.WriteString(generateStatement(s, "    ", device.Name))
		}
		out.WriteString("\n")
	}

	if device.Outputs != nil {
		out.WriteString("    // Write outputs\n")
		for _, o := range device.Outputs.Outputs {
			fmt.Fprintf(&out, "    write_actuator_%s(state);\n", o.Pin)
		}
	}

	out.WriteString("}\n\n")

	if emitMain {
		out.WriteString("int main(void) {\n")
		if device.Boot != nil {
			fmt.Fprintf(&out, "    %s_boot();\n", device.Name)
		}
		out.WriteString("    while (1) {\n")
		fmt.Fprintf(&out, "        %s_loop();\n", device.Name)
		fmt.Fprintf(&out, "        delay_ms(CYCLE_MS);\n")
		out.WriteString("    }\n")
		out.WriteString("    return 0;\n")
		out.WriteString("}\n")
	}

	return out.String()
}

func generateStatement(node ast.Node, indent string, deviceName string) string {
	var out strings.Builder

	upper := strings.ToUpper(deviceName)

	switch stmt := node.(type) {
	case *ast.AssignStatement:
		// if the value is a state name, prefix it with DEVICE_
		val := stmt.Value
		if isUpperIdent(val) {
			val = upper + "_" + val
		}
		fmt.Fprintf(&out, "%s%s = %s;\n", indent, stmt.Name, val)

	case *ast.OutputCall:
		state := "0"
		if stmt.State == "on" {
			state = "1"
		}
		fmt.Fprintf(&out, "%swrite_actuator_%s(%s);\n", indent, stmt.Relay, state)

	case *ast.IfStatement:
		right := stmt.Right
		if right != "true" && right != "false" {
			right = strings.ToUpper(right)
		}
		fmt.Fprintf(&out, "%sif (%s %s %s) {\n", indent, stmt.Left, stmt.Operator, right)
		fmt.Fprintf(&out, "%s    %s_failsafe();\n", indent, deviceName)
		fmt.Fprintf(&out, "%s    return;\n", indent)
		fmt.Fprintf(&out, "%s}\n", indent)

	case *ast.IfElseStatement:
		leftVal := stmt.LeftVal
		if leftVal != "true" && leftVal != "false" {
			leftVal = strings.ToUpper(leftVal)
		}
		rightVal := stmt.RightVal
		if rightVal != "true" && rightVal != "false" {
			rightVal = strings.ToUpper(rightVal)
		}
		fmt.Fprintf(&out, "%sif (%s %s %s && %s %s %s) {\n",
			indent,
			stmt.LeftVar, stmt.LeftOp, leftVal,
			stmt.RightVar, stmt.RightOp, rightVal)
		out.WriteString(generateStatement(stmt.Then, indent+"    ", deviceName))
		fmt.Fprintf(&out, "%s} else {\n", indent)
		out.WriteString(generateStatement(stmt.Else, indent+"    ", deviceName))
		fmt.Fprintf(&out, "%s}\n", indent)

	case *ast.IfOrStatement:
		fmt.Fprintf(&out, "%sif (%s %s %s || %s %s %s) {\n",
			indent,
			stmt.LeftVar, stmt.LeftOp, strings.ToUpper(stmt.LeftVal),
			stmt.RightVar, stmt.RightOp, strings.ToUpper(stmt.RightVal))
		fmt.Fprintf(&out, "%s    %s_failsafe();\n", indent, deviceName)
		fmt.Fprintf(&out, "%s    return;\n", indent)
		fmt.Fprintf(&out, "%s}\n", indent)

	case *ast.IfOrElseStatement:
		leftVal := stmt.LeftVal
		if leftVal != "true" && leftVal != "false" {
			leftVal = strings.ToUpper(leftVal)
		}
		rightVal := stmt.RightVal
		if rightVal != "true" && rightVal != "false" {
			rightVal = strings.ToUpper(rightVal)
		}
		fmt.Fprintf(&out, "%sif (%s %s %s || %s %s %s) {\n",
			indent,
			stmt.LeftVar, stmt.LeftOp, leftVal,
			stmt.RightVar, stmt.RightOp, rightVal)
		out.WriteString(generateStatement(stmt.Then, indent+"    ", deviceName))
		fmt.Fprintf(&out, "%s} else {\n", indent)
		out.WriteString(generateStatement(stmt.Else, indent+"    ", deviceName))
		fmt.Fprintf(&out, "%s}\n", indent)
	}

	return out.String()
}

// isUpperIdent returns true if a value looks like a state name (all caps, letters/underscores only)
func isUpperIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if !((c >= 'A' && c <= 'Z') || c == '_') {
			return false
		}
	}
	return true
}
