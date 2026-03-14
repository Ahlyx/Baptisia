#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include "hal.h"

#define WATCHDOG_MS 500
#define CYCLE_MS 100

// Constants
#define MAX_PSI 100
#define MAX_FLOW 1000

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
#define PSI_MAX 200
#define FLOW_MIN 0
#define FLOW_MAX 200

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

int main(void) {
    pump_boot();
    while (1) {
        pump_loop();
        delay_ms(CYCLE_MS);
    }
    return 0;
}
