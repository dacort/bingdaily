package main

import (
	// #cgo LDFLAGS: -framework CoreFoundation -framework IOKit
	// int CanSleep();
	// void WillWake();
	// void WillSleep();
	// #include "main.h"
	"C"
)

func registerNotification() {
	C.registerNotifications()
}
