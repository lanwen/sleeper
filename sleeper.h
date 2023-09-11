#ifndef sleeper_h
#define sleeper_h

#include <ctype.h>
#include <stdlib.h>
#include <stdio.h>

#include <mach/mach_port.h>
#include <mach/mach_interface.h>
#include <mach/mach_init.h>

#include <IOKit/pwr_mgt/IOPMLib.h>
#include <IOKit/IOMessage.h>

int registerNotifications();

void unregisterNotifications();

void Started();
void WillWake();
void WillSleep();

#endif /* sleeper_h */