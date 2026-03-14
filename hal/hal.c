#include "hal.h"
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

int32_t read_sensor_A0(void) {
    FILE *f = fopen("/tmp/baptisia_sensors.txt", "r");
    if (!f) return 0;
    int32_t psi, flow, temp;
    fscanf(f, "%d %d %d", &psi, &flow, &temp);
    fclose(f);
    return psi;
}

int32_t read_sensor_A1(void) {
    FILE *f = fopen("/tmp/baptisia_sensors.txt", "r");
    if (!f) return 0;
    int32_t psi, flow, temp;
    fscanf(f, "%d %d %d", &psi, &flow, &temp);
    fclose(f);
    return flow;
}

void write_actuator_D0(int state) {
    FILE *f = fopen("/tmp/baptisia_actuator.txt", "w");
    if (!f) return;
    fprintf(f, "%d\n", state ? 1 : 0);
    fclose(f);
}

void write_actuator_motor_relay(int state) {
    if (state == 0) {
        printf("[PUMP] state = FAULT — motor_relay OFF\n");
        FILE *f = fopen("/tmp/baptisia_actuator.txt", "w");
        if (!f) return;
        fprintf(f, "2\n");
        fclose(f);
    } else {
        printf("[PUMP] state = RUNNING — motor_relay ON\n");
    }
}

void watchdog_reset(void) {}

void delay_ms(int ms) {
    usleep(ms * 1000);
}
