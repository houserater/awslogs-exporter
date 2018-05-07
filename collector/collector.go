package collector

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"

	"github.com/houserater/awslogs-exporter/log"
	"github.com/houserater/awslogs-exporter/types"
)

const (
	namespace = "awslogs"
	timeout   = 10 * time.Second
)

// Metrics descriptions
var (
	// exporter metrics
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last query of AWS Logs successful.",
		[]string{"region"}, nil,
	)

	// Log group metrics
	logGroupCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logGroupCount"),
		"The total number of log groups",
		[]string{"region"}, nil,
	)

	logMessageCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logMessageCount"),
		"The total number of log messages within start time",
		[]string{"region", "group"}, nil,
	)

	logMessage = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "logMessage"),
		"A log event message",
		[]string{"region", "group", "message", "date"}, nil,
	)
)

// Exporter collects AWS Logs metrics
type Exporter struct {
	sync.Mutex                    // Our exporter object will be locakble to protect from concurrent scrapes
	client        AWSLogsGatherer // Custom AWS Logs client to get information from the log groups
	region        string          // The region where the exporter will scrape
	timeout       time.Duration   // The timeout for the whole gathering process
}

// New returns an initialized exporter
func New(awsRegion string, logStreamNamePrefix string, logHistory int64) (*Exporter, error) {
	c, err := NewAWSLogsClient(awsRegion, &logStreamNamePrefix, logHistory)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		Mutex:         sync.Mutex{},
		client:        c,
		region:        awsRegion,
		timeout:       timeout,
	}, nil

}

// sendSafeMetric uses context to cancel the send over a closed channel.
// If a main function finishes (for example due to to timeout), the goroutines running in background will
// try to send metrics over a closed channel, this will panic, this way the context will check first
// if the iteraiton has been finished and dont let continue sending the metric
func sendSafeMetric(ctx context.Context, ch chan<- prometheus.Metric, metric prometheus.Metric) error {
	// Check if iteration has finished
	select {
	case <-ctx.Done():
		log.Errorf("Tried to send a metric after collection context has finished, metric: %s", metric)
		return ctx.Err()
	default: // continue
	}
	// If no then send the metric
	ch <- metric
	return nil
}

// Describe describes all the metrics ever exported by the AWS Logs exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- logGroupCount
	ch <- logMessageCount
	ch <- logMessage
}

// Collect fetches the stats from configured AWS Logs and delivers them
// as Prometheus metrics. It implements prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	log.Debugf("Start collecting...")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	e.Lock()
	defer e.Unlock()

	// Get log groups
	cs, err := e.client.GetLogGroups()
	if err != nil {
		sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(up, prometheus.GaugeValue, 0, e.region))
		log.Errorf("Error collecting metrics: %v", err)
		return
	}

	e.collectLogGroupMetrics(ctx, ch, cs)

	// Start getting metrics per cluster on its own goroutine
	errC := make(chan bool)
	totalCs := 0 // total cluster metrics gorotine ran

	for _, c := range cs {
		totalCs++
		go func(c types.AWSLogGroup) {
			// Get services
			ss, err := e.client.GetLogEvents(&c)
			if err != nil {
				errC <- true
				log.Errorf("Error collecting log group stream metrics: %v", err)
				return
			}

			e.collectLogGroupStreamMetrics(ctx, ch, ss)

			errC <- false
		}(*c)
	}

	// Grab result or not result error for each goroutine, on first error exit
	result := float64(1)

ServiceCollector:
	for i := 0; i < totalCs; i++ {
		select {
		case err := <-errC:
			if err {
				result = 0
				break ServiceCollector
			}
		case <-time.After(e.timeout):
			log.Errorf("Error collecting metrics: Timeout making calls, waited for %v  without response", e.timeout)
			result = 0
			break ServiceCollector
		}

	}
	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, result, e.region,
	)
}

func (e *Exporter) collectLogGroupMetrics(ctx context.Context, ch chan<- prometheus.Metric, groups []*types.AWSLogGroup) {
	// Total log group count
	sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(logGroupCount, prometheus.GaugeValue, float64(len(groups)), e.region))
}

func (e *Exporter) collectLogGroupStreamMetrics(ctx context.Context, ch chan<- prometheus.Metric, events *types.AWSLogGroupEvents) {
	sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(logMessageCount, prometheus.GaugeValue, float64(len(events.Logs)), e.region, events.Group.Name))

	for _, event := range events.Logs {
		log := *event.Message
		date := time.Unix(0, *event.Timestamp * int64(time.Millisecond)).Format(time.RFC3339)

		sendSafeMetric(ctx, ch, prometheus.MustNewConstMetric(logMessage, prometheus.UntypedValue, 1, e.region, events.Group.Name, log, date))
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("awslogs_exporter"))
}
