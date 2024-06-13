package main

import (
	"log"
)

func LogInfo(format string, v ...any) {
	log.Printf(format, v)
}
