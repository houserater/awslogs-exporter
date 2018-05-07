package types

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

// AWSLogStream represents a collection of log statements on CloudWatch
type AWSLogGroupEvents struct {
	Group *AWSLogGroup                      // Log group reference
	Logs []*cloudwatchlogs.FilteredLogEvent // Log events
}

// AWSLogGroup represents a collection of streams on CloudWatch
type AWSLogGroup struct {
	ID   string // ARN of the log group
	Name string // Name of the log group
}
