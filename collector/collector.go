package collector

import (
	"regexp"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/version"

	"github.com/slok/ecs-exporter/log"
	"github.com/slok/ecs-exporter/types"
)

const (
	namespace = "ecs"
	timeout   = 10 * time.Second
)

// Metrics descriptions
var (
	// exporter metrics
	up = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "up"),
		"Was the last query of ecs successful.",
		[]string{"region"}, nil,
	)

	// Clusters metrics
	clusterCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "clusters"),
		"The total number of clusters",
		[]string{"region"}, nil,
	)

	//  Services metrics
	serviceCount = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "services"),
		"The total number of services",
		[]string{"region", "cluster"}, nil,
	)

	serviceDesired = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "service_desired_tasks"),
		"The desired number of instantiations of the task definition to keep running regarding a service",
		[]string{"region", "cluster", "service"}, nil,
	)

	servicePending = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "service_pending_tasks"),
		"The number of tasks in the cluster that are in the PENDING state regarding a service",
		[]string{"region", "cluster", "service"}, nil,
	)

	serviceRunning = prometheus.NewDesc(
		prometheus.BuildFQName(namespace, "", "service_running_tasks"),
		"The number of tasks in the cluster that are in the RUNNING state regarding a service",
		[]string{"region", "cluster", "service"}, nil,
	)
)

// Exporter collects ECS clusters metrics
type Exporter struct {
	sync.Mutex                   // Our exporter object will be locakble to protect from concurrent scrapes
	client        ECSGatherer    // Custom ECS client to get information from the clusters
	region        string         // The region where the exporter will scrape
	clusterFilter *regexp.Regexp // Compiled regular expresion to filter clusters
}

// New returns an initialized exporter
func New(awsRegion string, clusterFilterRegexp string) (*Exporter, error) {
	c, err := NewECSClient(awsRegion)
	if err != nil {
		return nil, err
	}

	cRegexp, err := regexp.Compile(clusterFilterRegexp)
	if err != nil {
		return nil, err
	}

	return &Exporter{
		Mutex:         sync.Mutex{},
		client:        c,
		region:        awsRegion,
		clusterFilter: cRegexp,
	}, nil

}

// Describe describes all the metrics ever exported by the ECS exporter. It
// implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- up
	ch <- clusterCount
}

// Collect fetches the stats from configured ECS and delivers them
// as Prometheus metrics. It implements prometheus.Collector
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	log.Debugf("Start collecting...")

	e.Lock()
	defer e.Unlock()

	// Get clusters
	cs, err := e.client.GetClusters()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(
			up, prometheus.GaugeValue, 0, e.region,
		)

		log.Errorf("Error collecting metrics: %v", err)
		return
	}

	e.collectClusterMetrics(ch, cs)

	// Start getting metrics per cluster on its own goroutine
	errC := make(chan bool)
	totalCs := 0 // total cluster metrics gorotine ran

	for _, c := range cs {
		// Filter not desired clusters
		if !e.validCluster(c) {
			log.Debugf("Cluster '%s' filtered", c.Name)
			continue
		}
		totalCs++
		go func(c types.ECSCluster) {
			// Get services
			ss, err := e.client.GetClusterServices(&c)
			if err != nil {
				errC <- true
				log.Errorf("Error collecting metrics: %v", err)
				return
			}
			e.collectClusterServicesMetrics(ch, &c, ss)
			errC <- false
		}(*c)
	}

	// Grab result or not result error for each goroutine, on first error exit
	result := float64(1)
	for i := 0; i < totalCs; i++ {
		select {
		case err := <-errC:
			if err {
				result = 0
				break
			}
		case <-time.After(timeout):
			log.Errorf("Error collecting metrics: Timeout making calls, waited for %v  without response", timeout)
			result = 0
		}

	}
	ch <- prometheus.MustNewConstMetric(
		up, prometheus.GaugeValue, result, e.region,
	)
}

// validCluster will return true if the cluster is valid for the exporter cluster filtering regexp, otherwise false
func (e *Exporter) validCluster(cluster *types.ECSCluster) bool {
	return e.clusterFilter.MatchString(cluster.Name)
}

func (e *Exporter) collectClusterMetrics(ch chan<- prometheus.Metric, clusters []*types.ECSCluster) {
	// Total cluster count
	ch <- prometheus.MustNewConstMetric(
		clusterCount, prometheus.GaugeValue, float64(len(clusters)), e.region,
	)
}

func (e *Exporter) collectClusterServicesMetrics(ch chan<- prometheus.Metric, cluster *types.ECSCluster, services []*types.ECSService) {

	// Total services
	ch <- prometheus.MustNewConstMetric(
		serviceCount, prometheus.GaugeValue, float64(len(services)), e.region, cluster.Name,
	)

	for _, s := range services {
		// Desired task count
		ch <- prometheus.MustNewConstMetric(
			serviceDesired, prometheus.GaugeValue, float64(s.DesiredT), e.region, cluster.Name, s.Name,
		)

		// Pending task count
		ch <- prometheus.MustNewConstMetric(
			servicePending, prometheus.GaugeValue, float64(s.PendingT), e.region, cluster.Name, s.Name,
		)

		// Running task count
		ch <- prometheus.MustNewConstMetric(
			serviceRunning, prometheus.GaugeValue, float64(s.RunningT), e.region, cluster.Name, s.Name,
		)
	}
}

func init() {
	prometheus.MustRegister(version.NewCollector("ecs_exporter"))
}
