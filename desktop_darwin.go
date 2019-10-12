package main

// To use the two libraries we need to define the respective flags, include the required header files and import "C" immediately after
import (
	// #cgo LDFLAGS: -framework CoreFoundation
	// #cgo LDFLAGS: -framework ApplicationServices
	// #include <CoreFoundation/CoreFoundation.h>
	// #include <ApplicationServices/ApplicationServices.h>
	//
	// extern void CoreDockSendNotification(CFStringRef /*notification*/, void * /*unknown*/);
	// int doit()
	// {
	// 	CoreDockSendNotification(CFSTR("com.apple.showdesktop.awake"), NULL);
	// 	return 0;
	// }
	"C"
	// other packages...
)

func showDesktop() {
	C.doit()
}
