# Baptisia

**A safety-enforcing domain-specific language for ICS/OT systems that compiles to C.**

Named after the wild indigo plant — hardy, resilient, and hard to kill.

---

## What It Is

Baptisia is a compiled language for programming industrial control systems (ICS), PLCs, and SCADA devices. It compiles `.ba` source files to C.

The core idea: safety-critical patterns are **enforced by the compiler**, not by convention. A Baptisia program physically cannot produce C output where control logic runs before safety checks, where a watchdog is missing, or where a failsafe is undefined. These are structural guarantees, not style guidelines.

---

## Why It Exists

Most PLCs in the field today are programmed in legacy languages designed in the 1980s and 90s. These tools have no concept of enforced execution order or mandatory safety interlocks. Safety is entirely convention-dependent — a tired engineer can write control logic that runs before checking safety conditions, and nothing in the toolchain stops them.

Stuxnet worked in part because safety logic on Siemens PLCs could be manipulated without the control system detecting the violation. This is a real and ongoing problem.

Baptisia enforces the correct execution order at the compiler level:

```
boot → read inputs → validate ranges → safety checks → control logic → write outputs
```

This order cannot be violated. The compiler will not generate code any other way.

---

## Language Syntax

### Basic device definition

```baptisia
device motor : PLC {
    watchdog: 500ms
    cycle: 100ms

    const {
        max_speed = 500
        max_temp = 100
    }

    vars {
        vol b state = false
        vol i32 speed = 0
        vol f32 temp = 0.0
    }

    boot {
        state = false
        output(motor_relay, off)
    }

    inputs {
        speed = val(sensor(A0), min: 0, max: 1000)
        temp  = val(sensor(A1), min: -40, max: 150)
    }

    outputs {
        motor_relay = actuator(D0)
    }

    safety {
        if speed >= max_speed : state = false
        if temp  >= max_temp  : state = false
    }

    failsafe {
        state = false
        output(motor_relay, off)
    }

    control {
        if speed < max_speed AND temp < max_temp : state = true
        else : state = false
    }
}
```

### Multi-state device with OR safety logic

```baptisia
device pump : SCADA {
    watchdog: 500ms
    cycle: 100ms

    const {
        max_psi  = 100
        max_flow = 500
    }

    states {
        IDLE
        RUNNING
        FAULT
        MAINTENANCE
    }

    vars {
        vol i32 psi  = 0
        vol i32 flow = 0
    }

    boot {
        state = IDLE
        output(motor_relay, off)
    }

    inputs {
        psi  = val(sensor(A0), min: 0, max: 500)
        flow = val(sensor(A1), min: 0, max: 500)
    }

    outputs {
        motor_relay = actuator(D0)
    }

    safety {
        if psi >= max_psi OR flow >= max_flow : state = FAULT
    }

    failsafe {
        state = FAULT
        output(motor_relay, off)
    }

    control {
        if psi < max_psi AND flow < max_flow : state = RUNNING
        else : state = IDLE
    }
}
```

### Language reference

| Keyword | Description |
|---|---|
| `vol` | Volatile — maps to C `volatile` |
| `b` | Boolean type (`bool`) |
| `i32` | 32-bit signed integer (`int32_t`) |
| `f32` | Single-precision float (`float`) |
| `const` | Named safety thresholds, emitted as `#define` |
| `states` | Named enum states for the device |
| `AND` | Compound condition — both sides must be true |
| `OR` | Compound condition — either side triggers |
| `//` | Single-line comment |

For the full language specification see [`docs/language-reference.md`](docs/language-reference.md).

---

## Compiler Pipeline

```
.ba file → Lexer → Tokens → Parser → AST → Semantic Analysis → Codegen → .c file
```

Written in Go. All stages are complete.

---

## Building

Requires Go 1.21+ and GCC.

```bash
git clone https://github.com/Ahlyx/Baptisia.git
cd Baptisia
go build ./...
```

---

## Usage

**Standard compile** — emits `main()`, produces a standalone C file:
```bash
go run main.go motor.ba
# produces hal/motor.c
```

**Simulation mode** — omits `main()`, for linking against the HAL and physics engine:
```bash
go run main.go -sim "test files/OR_logic.ba"
# produces hal/OR_logic.c without main()
```

---

## HAL and Physics Simulation

Baptisia includes a simulation stack that lets generated C run on Linux without real hardware. The physics engine models realistic pump behavior — pressure builds based on flow rate and pipe resistance, temperature rises with run time, and faults latch until manually reset.

### Architecture

```
.ba file → compiler → C control loop
                            ↓
                       HAL layer (hal/hal.c)
                            ↓
                    Physics engine (physics/)
                    models pressure, flow, temp
                            ↓
                    Sensor values fed back
                    into control loop each cycle
```

### Running the full simulation

```bash
# 1. Compile pump in sim mode
go run main.go -sim "test files/OR_logic.ba"

# 2. Build the C control binary
gcc -o hal/pump_control hal/OR_logic.c hal/hal.c hal/sim.c -I hal/

# 3. Run the full pipeline
go run cmd/fullsim/main.go
```

### Expected output

```
=== Baptisia Full Simulation ===
Physics engine + compiled Baptisia control loop

[CYCLE 0]  psi=0   flow=0   temp=20 state=RUNNING
[CYCLE 1]  psi=0   flow=8   temp=20 state=RUNNING
...
[CYCLE 55] psi=97  flow=400 temp=36 state=RUNNING
[CYCLE 56] psi=100 flow=400 temp=36 state=FAULT
[CYCLE 57] psi=94  flow=380 temp=36 state=FAULT
...
[CYCLE 76] psi=0   flow=0   temp=31 state=FAULT

=== Simulation Complete ===
Total cycles: 80
Final state:  FAULT
Fault latched: YES — manual reset required
Final PSI:    0
Final flow:   0
Final temp:   30°C
```

At cycle 56, PSI reaches `MAX_PSI`. The OR safety condition fires, `pump_failsafe()` is called, and the loop returns before control logic can run. The fault latches — the pump cannot return to RUNNING without a manual reset. Pressure and flow bleed off naturally over the following cycles, modeling hydraulic inertia.

### Standalone physics test

To test the physics model without the compiled control loop:

```bash
go run cmd/simulate/main.go
```

---

## Safety Guarantees

The compiler enforces these properties on every generated file:

- **Watchdog reset** at the top of every cycle, before anything else runs
- **Input validation** — sensor reads are range-checked before safety or control logic executes
- **Safety before control** — safety checks always run first; if violated, failsafe is called and the function returns
- **Symmetric failsafe** — failsafe and boot always perform the same state initialization
- **No fallthrough** — safety violations emit `return` after calling failsafe, preventing control logic from overriding a fault state
- **Undefined identifier rejection** — the semantic analyzer catches references to undeclared variables, constants, and state names before any C is emitted

These are not conventions. They are structural properties of the compiler output.

---

## Project Structure

```
Baptisia/
  main.go                   # CLI entry point, AST printer
  go.mod
  lexer/
    token.go                # Token type definitions
    lexer.go                # Lexer
  ast/
    nodes.go                # AST node definitions
  parser/
    parser.go               # Parser
  semantic/
    analyzer.go             # Semantic analysis — block validation, undefined identifier checks
  codegen/
    codegen.go              # C code generator
  hal/
    hal.h                   # HAL interface
    hal.c                   # Simulated sensor/actuator implementation
    sim.c                   # Single-cycle C entry point for fullsim
  physics/
    physics.go              # Pump physics model (pressure, flow, temperature)
    engine.go               # Simulation engine, manages cycles and state
  cmd/
    simulate/main.go        # Standalone physics test
    fullsim/main.go         # Full pipeline — physics engine + compiled control loop
  docs/
    language-reference.md   # Full language specification
  test files/
    motor.ba                # Motor controller (PLC target)
    OR_logic.ba             # Pump controller with OR safety and enum states
```

## License

MIT — see `LICENSE`.

Built by [@Ahlyx](https://github.com/Ahlyx)
