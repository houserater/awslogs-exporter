package collector

import (
	"fmt"
	"time"
	"sort"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"

	"github.com/houserater/awslogs-exporter/log"
	"github.com/houserater/awslogs-exporter/types"
)

// AWSLogsGatherer is the interface that implements the methods required to gather ECS data
type AWSLogsGatherer interface {
	GetLogGroups() ([]*types.AWSLogGroup, error)
	GetLogEvents(group *types.AWSLogGroup) (*types.AWSLogGroupEvents, error)
}

// Generate ECS API mocks running go generate
//go:generate mockgen -source ../vendor/github.com/aws/aws-sdk-go/service/ecs/ecsiface/interface.go -package sdk -destination ../mock/aws/sdk/ecsiface_mock.go

// AWSLogsClient is a wrapper for AWS CloudWatch client that implements helpers to get logging metrics
type AWSLogsClient struct {
	client             cloudwatchlogsiface.CloudWatchLogsAPI
	logGroupNamePrefix *string
	logHistory         int64
}

// NewECSClient will return an initialized ECSClient
func NewAWSLogsClient(awsRegion string, logGroupNamePrefix *string, logHistory int64) (*AWSLogsClient, error) {
	// Create AWS session
	s := session.New(&aws.Config{Region: aws.String(awsRegion)})
	if s == nil {
		return nil, fmt.Errorf("error creating aws session")
	}

	return &AWSLogsClient{
		client: cloudwatchlogs.New(s),
		logGroupNamePrefix: logGroupNamePrefix,
		logHistory: logHistory,
	}, nil
}

// GetClusters will get the clusters from the ECS API
func (e *AWSLogsClient) GetLogGroups() ([]*types.AWSLogGroup, error) {
	params := &cloudwatchlogs.DescribeLogGroupsInput{}
	if len(*e.logGroupNamePrefix) > 0 {
		params.LogGroupNamePrefix = e.logGroupNamePrefix
	}

	// Get log groups
	log.Debugf("Getting log groups for region")
	resp, err := e.client.DescribeLogGroups(params)
	if err != nil {
		return nil, err
	}

	cs := []*types.AWSLogGroup{}
	log.Debugf("Getting log group descriptions")
	for _, c := range resp.LogGroups {
		ec := &types.AWSLogGroup{
			ID:   aws.StringValue(c.Arn),
			Name: aws.StringValue(c.LogGroupName),
		}
		cs = append(cs, ec)
	}

	log.Debugf("Got %d log groups", len(cs))
	return cs, nil
}

// GetClusterServices will return all the services from a cluster
func (e *AWSLogsClient) GetLogEvents(group *types.AWSLogGroup) (*types.AWSLogGroupEvents, error) {
	startTime := (time.Now().Unix() - e.logHistory) * 1000

	params := &cloudwatchlogs.FilterLogEventsInput{
		LogGroupName: &group.Name,
		StartTime: &[]int64{startTime}[0],
	}

	// Get log messages
	log.Debugf("Getting log events for region")
	resp, err := e.client.FilterLogEvents(params)
	if err != nil {
		return nil, err
	}

	sortedEvents := make([]*cloudwatchlogs.FilteredLogEvent, len(resp.Events))
	copy(sortedEvents, resp.Events)

	sort.Slice(sortedEvents, func(i, j int) bool {
		return *resp.Events[i].Timestamp > *resp.Events[j].Timestamp
	})

	log.Debugf("Got %d log events", len(resp.Events))
	return &types.AWSLogGroupEvents{
		Group: group,
		Logs: sortedEvents,
	}, nil
}
