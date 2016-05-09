package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"os"
)

func createQueue(svc *sqs.SQS) (string, error) {
	params := &sqs.CreateQueueInput{
		QueueName: aws.String("my-queue"),
	}
	resp, err := svc.CreateQueue(params)
	if err != nil {
		return "", err
	}

	return *resp.QueueUrl, err
}

func getQueueArn(queueURL string, svc *sqs.SQS) (string, error) {
	params := &sqs.GetQueueAttributesInput{
		QueueUrl: aws.String(queueURL),
		AttributeNames: []*string{
			aws.String("QueueArn"),
		},
	}
	resp, err := svc.GetQueueAttributes(params)

	if err != nil {
		return "", err
	}

	return *resp.Attributes["QueueArn"], nil
}

func main() {
	svc := sqs.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})

	queueURL, err := createQueue(svc)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(queueURL)

	queueArn, err := getQueueArn(queueURL, svc)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println(queueArn)
}