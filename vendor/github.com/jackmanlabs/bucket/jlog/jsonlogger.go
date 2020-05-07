package jlog

import (
	"encoding/json"
	"log"
	"runtime"
	"strings"
)

/*
This orignally printed the data to stderr, like the log package by default, but
I decided it was better to let the user configure the log package, taking
advantage of the log package configuration.
*/
func Log(thing interface{}) {

	pc, file, line, _ := runtime.Caller(1)
	func_ := runtime.FuncForPC(pc)
	funcChunks := strings.Split(func_.Name(), "/")
	funcName := funcChunks[len(funcChunks)-1]

	log.Printf("%s:%d\t(%s)\n", file, line, funcName)

	if output, err := json.MarshalIndent(thing, "", "\t"); err == nil {
		log.Printf("%s\n", output)
	} else {
		log.Printf("%s\n", err.Error())
	}
}
