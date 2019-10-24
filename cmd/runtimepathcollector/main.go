package main

import (
	"encoding/json"
	"fmt"
	"github.com/fabric8-analytics/poc-ocp-upgrade-prediction/pkg/sqsconnect"
	"github.com/maruel/panicparse/stack"
	"go.uber.org/zap"
	"log"
	"os"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not initialize logging with zap. Panicking.")
		os.Exit(-1)
	}
	var allStacks stack.Callstacks

	for true {
		// Get messages from SQS.
		messageBatch := sqsconnect.GetStackTraceBatchFromQueue()
		if messageBatch == nil {
			logger.Sugar().Errorf("Got error while retrieving messages.")
			os.Exit(-2)
		}
		if len(messageBatch) == 0 {
			logger.Sugar().Info("Got no messages, exiting.")
			break
		}
		logger.Sugar().Info("Got %d messages\n", len(messageBatch))

		// Filter duplicate stacktraces from the messages, maintain stack queue.
		for _, message := range messageBatch {
			var buffer stack.Context
			err := json.Unmarshal([]byte(message), &buffer)
			if err != nil {
				logger.Sugar().Errorf("Failed to Unmarshal message, got error: %v\n", err)
				os.Exit(-3)
			}
			allStacks = stack.AggregateSubsets(buffer.Goroutines, allStacks)
			logger.Sugar().Infof("Stack count: %d\n", len(allStacks))
		}
	}
	logger.Sugar().Infof("After filtration, unique stack count is: %d\n", len(allStacks))
	logger.Sugar().Infof("Writing unique stacks to file.")

	f, err := os.OpenFile(os.Getenv("OUTPUT_LOG"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Log to the output log.
	for _, curStack := range allStacks {
		// Write to log in JSON per line format, can be processed by Pandas.
		for _, trace := range *curStack {
			if _, err := f.Write([]byte(trace + "\n")); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Close the file.
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}
