package main

import (
	"fmt"
)

func main() {
	fmt.Printf("[hello log]\n")
	var log Log
	err := log.Init()
	if err != nil {
		return
	}
	log, err = NewLog()
	if err != nil {
		return
	}
	fmt.Printf("[log = %#v]\n", log)
	//log.Write(LOG_DEBUG, "hellvo lua:\n")
	//log.Write(LOG_DEBUG, "hellvo lua:%s\n", "sdjfld")
	log.Write(LOG_DEBUG, "hellvo lua:")
	log.Write(LOG_DEBUG, "hellvo lua:%s", "sdjfld")
	log.Flush()
	return
}
