package main

import (
	"bufio"
	"encoding/json"
	"github.com/maruel/panicparse/stack"
	"go.uber.org/zap"
	"io"
	"os"
)

func main() {
	logger, err := zap.NewDevelopment()
	// First collect all the function call messages in a log.
	//sqsconnect.StoreFunctionCallsToLog()

	fPtr, err := os.Open(os.Getenv("RUNTIME_LOG_FILENAME"))
	if err != nil {
		logger.Sugar().Errorf("Could not open log file for reading, got error: %v\n", err)
	}
	rdr := bufio.NewReader(fPtr)

	for callStackStr, err := rdr.ReadString('\n'); callStackStr != "" && err == nil; callStackStr, err = rdr.ReadString('\n') {
		var callStack stack.Context
		err := json.Unmarshal([]byte(callStackStr), &callStack)
		if err != nil {
			logger.Sugar().Errorf("Could not parse callStack log, got error: %v\n", err)
		}
		stack.Aggregate(callStack.Goroutines, stack.AnyPointer)
		logger.Sugar().Infof("%#v\n", callStack)
	}
	// Got a real error
	if err != nil && err != io.EOF {
		logger.Sugar().Errorf("Got error: %v\n", err)
	}
}
