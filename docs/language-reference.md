# Baptisia Language Reference

Version 0.2 — covers all features currently implemented in the compiler.

---

## Overview

A Baptisia program describes a single physical device — a PLC, SCADA controller, or embedded system. The compiler takes a `.ba` source file and produces a `.c` file that implements the device's control loop with guaranteed safety properties.

Every compiled device has the same execution structure, enforced by the compiler:

```
1. boot()       — runs once on startup, initializes state
2. loop() {
     watchdog_reset()
     read inputs + validate ranges
     safety checks
     control logic
     write outputs
   }
```

This order is not configurable. The compiler will not generate code that deviates from it.

---

## File Structure

A `.ba` file contains exactly one `device` block. The device block contains named sub-blocks that define each aspect of the device's behavior. Blocks can appear in any order inside the device — the compiler enforces execution order, not source order.

```baptisia
device <n> : <target> {
    watchdog: <time>
    cycle: <time>

    const { ... }
    states { ... }
    vars { ... }
    boot { ... }
    inputs { ... }
    outputs { ... }
    safety { ... }
    failsafe { ... }
    control { ... }
}
```

---

## Device Declaration

```baptisia
device motor : PLC {
    ...
}
```

`device` is the top-level keyword. The name (`motor`) becomes the prefix for all generated C functions — `motor_boot()`, `motor_loop()`, `motor_failsafe()`. The target (`PLC`, `SCADA`) is metadata recorded in the AST and currently used as a label.

---

## Timing

```baptisia
watchdog: 500ms
cycle: 100ms
```

Both are required. The compiler emits them as C `#define` constants:

```c
#define WATCHDOG_MS 500
#define CYCLE_MS 100
```

`watchdog` is the maximum time allowed between watchdog resets. If the control loop hangs, the watchdog timer expires and resets the device. `cycle` is the target loop period — `delay_ms(CYCLE_MS)` is called at the end of every cycle in `main()`.

Time values are integers suffixed with `ms`. No other time unit is currently supported.

---

## Constants Block

```baptisia
const {
    max_speed = 500
    max_temp = 100
}
```

Constants define named safety thresholds. They are emitted at the top of the C file as `#define` macros, uppercased:

```c
#define MAX_SPEED 500
#define MAX_TEMP 100
```

Constants must be referenced by name in safety and control conditions. Using raw literals in conditions is valid but defeats the purpose — named constants make the safety thresholds visible and auditable.

Constant names are lowercase in `.ba` files and automatically uppercased in C output.

---

## States Block

```baptisia
states {
    IDLE
    RUNNING
    FAULT
    MAINTENANCE
}
```

The `states` block defines named enum states for the device. The compiler emits a `typedef enum` in C, prefixed with the device name:

```c
typedef enum {
    PUMP_IDLE,
    PUMP_RUNNING,
    PUMP_FAULT,
    PUMP_MAINTENANCE
} pump_state_t;

volatile pump_state_t state = PUMP_IDLE;
```

The first state listed becomes the initial value. State names must be all uppercase. If a `states` block is present, the `state` variable is automatically declared with the correct enum type — do not declare it manually in `vars`.

State names are used directly in assignments throughout the device:

```baptisia
state = FAULT
state = RUNNING
state = IDLE
```

The compiler detects uppercase identifiers and emits the correct `DEVICE_STATENAME` prefix automatically.

---

## Variables Block

```baptisia
vars {
    vol b state = false
    vol i32 speed = 0
    vol f32 temp = 0.0
}
```

Each variable declaration has the form:

```
[vol] <type> <name> = <value>
```

`vol` marks the variable as volatile, which maps to C `volatile`. All hardware-mapped variables should be volatile.

### Primitive types

| Baptisia type | C type | Description |
|---|---|---|
| `b` | `bool` | Boolean — `true` or `false` |
| `i32` | `int32_t` | 32-bit signed integer |
| `f32` | `float` | Single-precision floating point |

If a `states` block is present, do not declare a `state` variable in `vars` — the compiler generates it from the enum definition.

---

## Boot Block

```baptisia
boot {
    state = false
    output(motor_relay, off)
}
```

The boot block runs once on startup before the control loop begins. It initializes state and sets outputs to their safe default values. The compiler emits it as a `<device>_boot()` function called once from `main()`.

Boot must establish the same safe condition as failsafe — they should be symmetric. If failsafe sets `state = FAULT` and turns the relay off, boot should initialize to a safe state and turn the relay off too.

Statements in boot can be assignments or output calls. If-statements are not currently supported in boot.

---

## Inputs Block

```baptisia
inputs {
    speed = val(sensor(A0), min: 0, max: 1000)
    temp  = val(sensor(A1), min: -40, max: 150)
}
```

Each input declaration maps a variable to a physical sensor pin with a validated range. The syntax is:

```
<variable> = val(sensor(<pin>), min: <value>, max: <value>)
```

The compiler emits:
- A `#define` pair for the sensor's valid range (`SPEED_MIN`, `SPEED_MAX`)
- A sensor read call inside `<device>_loop()`
- An immediate range check — if the reading is outside the valid range, `<device>_failsafe()` is called and the loop returns

```c
speed = read_sensor_A0();
if (speed < SPEED_MIN || speed > SPEED_MAX) { motor_failsafe(); return; }
```

Input validation runs before safety checks. A sensor reading outside its valid range is treated the same as a safety violation — the failsafe fires and the cycle ends.

Pin names are passed through to the HAL function name: `sensor(A0)` generates `read_sensor_A0()`.

---

## Outputs Block

```baptisia
outputs {
    motor_relay = actuator(D0)
}
```

Each output declaration maps a named relay to a physical actuator pin. The syntax is:

```
<name> = actuator(<pin>)
```

The compiler emits a write call at the end of each loop cycle:

```c
write_actuator_D0(state);
```

Output writes happen last in the loop, after safety and control logic have run. Pin names are passed through to the HAL function name: `actuator(D0)` generates `write_actuator_D0()`.

---

## Safety Block

```baptisia
safety {
    if speed >= max_speed : state = false
    if temp  >= max_temp  : state = false
}
```

The safety block contains the device's safety interlocks. It runs before control logic on every cycle. Each condition is evaluated independently.

When a safety condition is triggered, the compiler emits a failsafe call followed by `return`:

```c
if (speed >= MAX_SPEED) {
    motor_failsafe();
    return;
}
```

This means control logic cannot run on the same cycle as a safety violation. The return is enforced by the compiler — it is not optional.

### Simple if

```baptisia
if <var> <op> <value> : state = <result>
```

Used for single-condition safety checks. The assignment on the right side of `:` is ignored — safety conditions always trigger the failsafe, not an inline assignment. The compiler generates `<device>_failsafe()` regardless of what the assignment says.

### OR compound condition

```baptisia
if psi >= max_psi OR flow >= max_flow : state = FAULT
```

Either condition triggering calls the failsafe. Emits `||` in C. Used when multiple independent hazards should trigger the same response.

### Supported operators

| Operator | Description |
|---|---|
| `>=` | Greater than or equal |
| `>` | Greater than |
| `<=` | Less than or equal |
| `<` | Less than |
| `==` | Equal |
| `!=` | Not equal |

All six operators are supported in safety, control, and compound conditions.

---

## Failsafe Block

```baptisia
failsafe {
    state = FAULT
    output(motor_relay, off)
}
```

The failsafe block defines the device's emergency shutdown procedure. It is emitted as a standalone `<device>_failsafe()` function that is called by safety violations, input range violations, and the boot sequence.

Failsafe should always set the device to a known safe state and de-energize all outputs. It must be symmetric with the boot block.

Statements in failsafe can be assignments or output calls.

### Output call

```baptisia
output(motor_relay, off)
```

Writes a value to a named relay. The relay name must match a declaration in the `outputs` block. Valid states are `off` (emits `0`) and `on` (emits `1`).

---

## Control Block

```baptisia
control {
    if speed < max_speed AND temp < max_temp : state = true
    else : state = false
}
```

The control block contains the device's normal operating logic. It runs after safety checks on every cycle, only if no safety condition was triggered.

### AND compound condition with else

```baptisia
if <var> <op> <value> AND <var> <op> <value> : <assignment>
else : <assignment>
```

Both conditions must be true for the `then` branch to execute. If either is false, the `else` branch executes. Emits `&&` in C. The `else` branch is required for AND conditions.

### OR compound condition with else

```baptisia
if <var> <op> <value> OR <var> <op> <value> : <assignment>
else : <assignment>
```

Either condition being true executes the `then` branch. Emits `||` in C. The `else` branch is optional.

---

## Semantic Analysis

The compiler performs semantic validation after parsing and before code generation. If any semantic error is found, no C output is produced.

Checks currently enforced:

- All mandatory blocks are present (`boot`, `inputs`, `outputs`, `safety`, `failsafe`, `control`)
- Every identifier referenced in `boot`, `safety`, and `control` is declared in `vars`, `consts`, `states`, or `outputs`
- Built-in literals (`true`, `false`, `on`, `off`) are always valid

A device that references an undeclared name will be rejected with a clear error message identifying the block and the unknown identifier.

---

## Comments

```baptisia
// this is a comment
```

Single-line comments only. Everything after `//` to the end of the line is ignored by the lexer. Block comments are not currently supported.

---

## Generated C Structure

For a device named `pump`, the compiler produces the following C structure:

```c
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include "hal.h"

#define WATCHDOG_MS 500
#define CYCLE_MS 100

// Constants
#define MAX_PSI 100
#define MAX_FLOW 500

// Enum state type
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

// Boot — runs once
void pump_boot(void) { ... }

// Failsafe — called on violation
void pump_failsafe(void) { ... }

// Control loop — runs every cycle
void pump_loop(void) {
    watchdog_reset();
    // read + validate inputs
    // safety checks
    // control logic
    // write outputs
}

int main(void) {
    pump_boot();
    while (1) {
        pump_loop();
        delay_ms(CYCLE_MS);
    }
    return 0;
}
```

---

## HAL Interface

The generated C file calls these HAL functions which must be implemented for your target:

| Function | Description |
|---|---|
| `read_sensor_<pin>()` | Returns current sensor reading as `int32_t` |
| `write_actuator_<pin>(int state)` | Sets actuator output |
| `watchdog_reset()` | Resets the hardware watchdog timer |
| `delay_ms(int ms)` | Blocks for the specified number of milliseconds |

A simulation HAL for Linux is provided in `hal/hal.c`. For real hardware, implement these functions against your platform's SDK.

---

## Compiler Flags

```bash
go run main.go [-sim] <file.ba>
```

| Flag | Description |
|---|---|
| `-sim` | Omit `main()` from output — use when linking against an external simulation driver |

Without `-sim`, the compiler emits a complete standalone C program including `main()`.

---

## Current Limitations

- One device per file
- No nested conditions (`if a AND b OR c`)
- No loops or iteration inside blocks
- Safety conditions trigger failsafe only — no inline recovery logic
- Single-line comments only
- No error recovery — parser prints error and continues, which can produce unexpected output
- Line numbers in error messages are present in the lexer but not yet propagated to all error sites
- No type checking — comparing an `f32` variable against an integer constant is not yet flagged

These are known limitations planned for future versions.
