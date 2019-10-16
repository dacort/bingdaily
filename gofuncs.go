package main

import (
	"C"
	"log"
)

//export CanSleep
func CanSleep() C.int {
	return 1
}

//export WillWake
func WillWake() {
	log.Printf("Will Wake, triggering thing")
	syncWithBing()
	log.Printf("woke")
}

//export WillSleep
func WillSleep() {
	log.Printf("Will Sleep")
}
