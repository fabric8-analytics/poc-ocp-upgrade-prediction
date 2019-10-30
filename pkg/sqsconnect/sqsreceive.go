package sqsconnect

import (
	"log"
	"os"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// GetStackTraceBatchFromQueue queries SQS to get a batch of upto 10 messages, each containing a stacktrace.
func GetStackTraceBatchFromQueue() []string {
	var messages []string
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
		Config: aws.Config{
			Region: aws.String(os.Getenv("AWS_SQS_REGION")),
			Endpoint: aws.String(os.Getenv("SQS_ENDPOINT")),
		},
	}))

	svc := sqs.New(sess)

	// URL to our queue
	qname := os.Getenv("AWS_SQS_QUEUE_NAME")
	qURL, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &qname,
	})
	if err != nil {
		sugarLogger.Errorf("Could not get queue URL, got error: %v\n", err)
		return nil
	}


	if qAttrs, err := svc.GetQueueAttributes(&sqs.GetQueueAttributesInput{
		AttributeNames: []*string{aws.String("ApproximateNumberOfMessages")},
		QueueUrl:       qURL.QueueUrl,
	}); err != nil || qAttrs.Attributes["ApproximateNumberOfMessages"] != aws.String("0") {
		if err != nil {
			sugarLogger.Errorf("Could not get queue attributes, got error: %v\n", err)
			return nil
		}
		// Get a batch of messages.
		result, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			AttributeNames: []*string{
				aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
			},
			MessageAttributeNames: []*string{
				aws.String(sqs.QueueAttributeNameAll),
			},
			QueueUrl:            qURL.QueueUrl,
			MaxNumberOfMessages: aws.Int64(10),
			VisibilityTimeout:   aws.Int64(60), // 60 seconds
			WaitTimeSeconds:     aws.Int64(0),
		})
		if err != nil {
			sugarLogger.Errorf("Got Error when receiving messages: %v\n", err)
			return nil
		}
		if len(result.Messages) == 0 {
			sugarLogger.Info("Received no messages")
			return []string{}
		} else {
			sugarLogger.Infof("Received %d messages.\n", len(result.Messages))

			for _, message := range result.Messages {
				// For debug purposes.
				if debug, _ := strconv.ParseBool(os.Getenv("DEBUG_SQS")); debug {
					writeMessagesToLog(message.Body)
				}
				messages = append(messages, *message.Body)
			}

			return messages
		}
	}
	sugarLogger.Info("No messages to process in queue at this time.")
	return []string{}
}

func writeMessagesToLog(messageBody *string) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(os.Getenv("RUNTIME_LOG_FILENAME"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	// Write to log in JSON per line format, can be processed by Pandas.
	if _, err := f.Write([]byte(*messageBody + "\n")); err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}
}