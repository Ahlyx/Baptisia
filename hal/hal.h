#ifndef HAL_H
#define HAL_H

#include <stdbool.h>
#include <stdint.h>

int32_t read_sensor_A0(void);
int32_t read_sensor_A1(void);
void write_actuator_D0(int state);
void write_actuator_motor_relay(int state);
void watchdog_reset(void);
void delay_ms(int ms);

#endif
