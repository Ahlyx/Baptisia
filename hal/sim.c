#include <stdio.h>
#include <stdint.h>

void pump_boot(void);
void pump_loop(void);

int main(void) {
    pump_loop();
    return 0;
}
