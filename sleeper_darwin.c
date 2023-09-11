#include "sleeper.h"

// a reference to the Root Power Domain IOService
io_connect_t root_port;
// notification port allocated by IORegisterForSystemPower
IONotificationPortRef notifyPortRef;
// notifier object, used to deregister later
io_object_t notifierObject;
// main run loop
CFRunLoopRef runLoop;

// https://developer.apple.com/documentation/corefoundation/1541546-cfrunloopobservercreate?language=objc
CFRunLoopObserverRef observer;

void sleepCallback(void *refCon, io_service_t service, natural_t messageType, void *messageArgument) {
    switch (messageType) {
        case kIOMessageCanSystemSleep:
            /*
                Idle sleep is about to kick in. This message will not be sent for forced sleep.
                Applications have a chance to prevent sleep by calling IOCancelPowerChange.
                Most applications should not prevent idle sleep.
                Power Management waits up to 30 seconds for you to either allow or deny idle
                sleep. If you don't acknowledge this power change by calling either
                IOAllowPowerChange or IOCancelPowerChange, the system will wait 30
                seconds then go to sleep.
            */
            IOAllowPowerChange(root_port, (long)messageArgument);
            break;
        case kIOMessageSystemWillSleep:
            /*
                The system WILL go to sleep. If you do not call IOAllowPowerChange or
                IOCancelPowerChange to acknowledge this message, sleep will be
                delayed by 30 seconds.
                NOTE: If you call IOCancelPowerChange to deny sleep it returns
                kIOReturnSuccess, however the system WILL still go to sleep.
            */
            WillSleep();
            IOAllowPowerChange(root_port, (long)messageArgument);
            break;
        case kIOMessageSystemWillPowerOn:
            // System has started the wake up process...
            WillWake();
            break;
        case kIOMessageSystemHasPoweredOn:
            // System has finished waking up...
            break;
        default:
            break;
    }
}

// This callback is called by observer on a specific stage of run loop
void runLoopObserverCallback(CFRunLoopObserverRef observer, CFRunLoopActivity activity, void *info) {
    switch (activity) {
        case kCFRunLoopBeforeSources:
            Started();
            break;
    }
}

int registerNotifications() {
    // register to receive system sleep notifications
    root_port = IORegisterForSystemPower(NULL, &notifyPortRef, sleepCallback, &notifierObject);
    if (root_port == 0) {
        return 1;
    }

    // create the observer which will fire a callback once the run loop is about to start processing power events
    observer = CFRunLoopObserverCreate(kCFAllocatorDefault, kCFRunLoopBeforeSources, false, 0, runLoopObserverCallback, NULL);
    // save that to shutdown later
    runLoop = CFRunLoopGetCurrent();

    // add the observer to the application runloop
    CFRunLoopAddObserver(runLoop, observer, kCFRunLoopCommonModes);
    // add the notification port to the application runloop
    CFRunLoopAddSource(runLoop, IONotificationPortGetRunLoopSource(notifyPortRef), kCFRunLoopCommonModes);
    /*
        Start the run loop to receive sleep notifications. Don't call CFRunLoopRun if this code
        is running on the main thread of a Cocoa or Carbon application. Cocoa and Carbon
        manage the main thread's run loop for you as part of their event handling
        mechanisms.
    */
    CFRunLoopRun();

    // will get there only once we stop the run loop
    return 0;
}

void unregisterNotifications() {
    CFRunLoopRemoveObserver(runLoop, observer, kCFRunLoopCommonModes);
    CFRelease(observer);

    // remove the sleep notification port from the application run loop
    CFRunLoopRemoveSource(runLoop, IONotificationPortGetRunLoopSource(notifyPortRef), kCFRunLoopCommonModes);

    // deregister for system sleep notifications
    IODeregisterForSystemPower(&notifierObject);

    // IORegisterForSystemPower implicitly opens the Root Power Domain IOService, so we close it here
    IOServiceClose(root_port);

    // destroy the notification port allocated by IORegisterForSystemPower
    IONotificationPortDestroy(notifyPortRef);

    CFRunLoopStop(runLoop);
}