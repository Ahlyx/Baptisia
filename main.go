package main

import (
	"baptisia/ast"
	"baptisia/codegen"
	"baptisia/lexer"
	"baptisia/parser"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	simMode := flag.Bool("sim", false, "omit main() for HAL simulation builds")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: baptisia [-sim] <filename>")
		os.Exit(1)
	}

	source, err := os.ReadFile(args[0])
	if err != nil {
		fmt.Println("Error reading file:", err)
		os.Exit(1)
	}

	l := lexer.New(string(source))
	p := parser.New(l)
	program := p.ParseProgram()

	// exit cleanly if parse errors were found
	if len(p.Errors) > 0 {
		fmt.Printf("Baptisia compiler errors in %s:\n\n", args[0])
		for _, e := range p.Errors {
			fmt.Printf("  error: %s\n", e)
		}
		fmt.Printf("\n%d error(s) found. No output generated.\n", len(p.Errors))
		os.Exit(1)
	}

	printAST(program.Device)

	fmt.Println("\n--- Generated C ---")
	cCode := codegen.Generate(program.Device, !*simMode)
	fmt.Println(cCode)

	base := filepath.Clean(filepath.Base(args[0]))
	outputName := "hal/" + strings.TrimSuffix(base, filepath.Ext(base)) + ".c"

	err = os.WriteFile(outputName, []byte(cCode), 0644)
	if err != nil {
		fmt.Println("Error writing output:", err)
		os.Exit(1)
	}
	fmt.Println("Written to", outputName)
}

func printAST(device *ast.DeviceNode) {
	fmt.Printf("Device: %s : %s\n", device.Name, device.Target)
	fmt.Printf("  Watchdog: %s\n", device.Watchdog)
	fmt.Printf("  Cycle: %s\n", device.Cycle)

	if device.Vars != nil {
		fmt.Println("  Vars:")
		for _, v := range device.Vars.Vars {
			fmt.Printf("    vol=%v type=%s name=%s value=%s\n",
				v.Volatile, v.TypeName, v.Name, v.Value)
		}
	}

	if device.Inputs != nil {
		fmt.Println("  Inputs:")
		for _, s := range device.Inputs.Inputs {
			fmt.Printf("    %s = sensor(%s) min=%s max=%s\n",
				s.Name, s.Pin, s.Min, s.Max)
		}
	}

	if device.Outputs != nil {
		fmt.Println("  Outputs:")
		for _, o := range device.Outputs.Outputs {
			fmt.Printf("    %s = actuator(%s)\n", o.Name, o.Pin)
		}
	}

	fmt.Println("  Safety:")
	if device.Safety != nil {
		for _, s := range device.Safety.Statements {
			switch stmt := s.(type) {
			case *ast.IfStatement:
				then := stmt.Then.(*ast.AssignStatement)
				fmt.Printf("    if %s %s %s : %s = %s\n",
					stmt.Left, stmt.Operator, stmt.Right,
					then.Name, then.Value)
			case *ast.IfOrStatement:
				then := stmt.Then.(*ast.AssignStatement)
				fmt.Printf("    if %s %s %s OR %s %s %s : %s = %s\n",
					stmt.LeftVar, stmt.LeftOp, stmt.LeftVal,
					stmt.RightVar, stmt.RightOp, stmt.RightVal,
					then.Name, then.Value)
			}
		}
	}

	fmt.Println("  Failsafe:")
	if device.Failsafe != nil {
		for _, s := range device.Failsafe.Statements {
			switch stmt := s.(type) {
			case *ast.AssignStatement:
				fmt.Printf("    %s = %s\n", stmt.Name, stmt.Value)
			case *ast.OutputCall:
				fmt.Printf("    output(%s, %s)\n", stmt.Relay, stmt.State)
			}
		}
	}

	fmt.Println("  Control:")
	if device.Control != nil {
		for _, s := range device.Control.Statements {
			switch stmt := s.(type) {
			case *ast.IfElseStatement:
				then := stmt.Then.(*ast.AssignStatement)
				els := stmt.Else.(*ast.AssignStatement)
				fmt.Printf("    if %s %s %s AND %s %s %s : %s = %s else : %s = %s\n",
					stmt.LeftVar, stmt.LeftOp, stmt.LeftVal,
					stmt.RightVar, stmt.RightOp, stmt.RightVal,
					then.Name, then.Value,
					els.Name, els.Value)
			}
		}
	}

	if device.Consts != nil {
		fmt.Println("  Constants:")
		for _, c := range device.Consts.Constants {
			fmt.Printf("    %s = %s\n", c.Name, c.Value)
		}
	}
}
