#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include "hal.h"

#define WATCHDOG_MS 500
#define CYCLE_MS 100

// Constants
#define MAX_SPEED 500
#define MAX_TEMP 100

// Variable declarations
volatile bool state = false;
volatile int32_t speed = 0;
volatile float temp = 0.0;

// Sensor validation ranges
#define SPEED_MIN 0
#define SPEED_MAX 1000
#define TEMP_MIN -40
#define TEMP_MAX 150

// Boot initialization - runs once on startup
void motor_boot(void) {
    state = false;
    write_actuator_motor_relay(0);
}

// Failsafe - called when safety conditions are violated
void motor_failsafe(void) {
    state = false;
    write_actuator_motor_relay(0);
}

// Main control loop for motor
void motor_loop(void) {
    watchdog_reset();

    // Read and validate inputs
    speed = read_sensor_A0();
    if (speed < SPEED_MIN || speed > SPEED_MAX) { motor_failsafe(); return; }
    temp = read_sensor_A1();
    if (temp < TEMP_MIN || temp > TEMP_MAX) { motor_failsafe(); return; }

    // Safety checks
    if (speed >= MAX_SPEED) {
        motor_failsafe();
    }
    if (temp >= MAX_TEMP) {
        motor_failsafe();
    }

    // Control logic
    if (speed < MAX_SPEED && temp < MAX_TEMP) {
        state = true;
    } else {
        state = false;
    }

    // Write outputs
    write_actuator_D0(state);
}

int main(void) {
    motor_boot();
    while (1) {
        motor_loop();
        delay_ms(CYCLE_MS);
    }
    return 0;
}
