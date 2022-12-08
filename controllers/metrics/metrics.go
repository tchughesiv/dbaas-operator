package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	dbaasv1beta1 "github.com/RHEcosystemAppEng/dbaas-operator/api/v1beta1"
)

const (
	// Metrics names.
	MetricNameDBaaSStackInstallationTotalDuration = "dbaas_stack_installation_total_duration_seconds"
	MetricNameDBaaSPlatformInstallationStatus     = "dbaas_platform_installation_status"
	MetricNameOperatorVersion                     = "dbaas_version_info"
	MetricNameDBaaSRequestsDurationSeconds        = "dbaas_requests_duration_seconds"
	MetricNameDBaaSRequestsErrorCount             = "dbaas_requests_error_count"

	// Metrics labels.
	MetricLabelName              = "name"
	MetricLabelStatus            = "status"
	MetricLabelVersion           = "version"
	MetricLabelProvider          = "provider"
	MetricLabelAccountName       = "account"
	MetricLabelNameSpace         = "namespace"
	MetricLabelInstanceID        = "instance_id"
	MetricLabelReason            = "reason"
	MetricLabelInstanceName      = "instance_name"
	MetricLabelCreationTimestamp = "creation_timestamp"
	MetricLabelConsoleULR        = "openshift_url"
	MetricLabelPlatformName      = "cloud_platform_name"
	MetricLabelResource          = "resource"
	MetricLabelEvent             = "event"
	MetricLabelErrorCd           = "error_cd"

	// Event label values
	LabelEventValueCreate = "create"
	LabelEventValueDelete = "delete"

	// Error Code label values
	LabelErrorCdValueResourceNotFound     = "resource_not_found"
	LabelErrorCdValueUnableToListPolicies = "unable_to_list_policies"

	// installationTimeStart base time == 0
	installationTimeStart = 0
	// installationTimeWidth is the width of a bucket in the histogram, here it is 1m
	installationTimeWidth = 60
	// installationTimeBuckets is the number of buckets, here it 10 minutes worth of 1m buckets
	installationTimeBuckets = 10
)

// DBaasPlatformInstallationGauge defines a gauge for DBaaSPlatformInstallationStatus
var DBaasPlatformInstallationGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameDBaaSPlatformInstallationStatus,
	Help: "The status of an installation of components and provider operators. values ( success=1, failed=0, in progress=2 ) ",
}, []string{MetricLabelName, MetricLabelStatus, MetricLabelVersion})

// DBaasStackInstallationHistogram defines a histogram for DBaasStackInstallation
var DBaasStackInstallationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: MetricNameDBaaSStackInstallationTotalDuration,
	Help: "Time in seconds installation of a DBaaS stack takes.",
	Buckets: prometheus.LinearBuckets(
		installationTimeStart,
		installationTimeWidth,
		installationTimeBuckets),
}, []string{MetricLabelVersion, MetricLabelCreationTimestamp})

// DBaasOperatorVersionInfo defines a gauge for DBaaS Operator version
var DBaasOperatorVersionInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: MetricNameOperatorVersion,
	Help: "The current version of DBaaS Operator installed in the cluster",
}, []string{MetricLabelVersion, MetricLabelConsoleULR, MetricLabelPlatformName})

// DBaaSRequestsDurationHistogram DBaaS Requests Duration Histogram for all DBaaS Resources
var DBaaSRequestsDurationHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name: MetricNameDBaaSRequestsDurationSeconds,
		Help: "Request durations histogram for given resource(e.g. inventory) and for a given event(e.g. create or delete)",
	},
	[]string{MetricLabelProvider, MetricLabelAccountName, MetricLabelNameSpace, MetricLabelResource, MetricLabelEvent})

// DBaaSRequestsErrorsCounter Total errors encountered counter
var DBaaSRequestsErrorsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: MetricNameDBaaSRequestsErrorCount,
		Help: "Total requests for a given resource(e.g. DBaaS Inventory), for a given event(e.g. create or delete), with a given error code (e.g. resource exists, resource not there)",
	},
	[]string{MetricLabelProvider, MetricLabelAccountName, MetricLabelNameSpace, MetricLabelResource, MetricLabelEvent, MetricLabelErrorCd})

// Execution tracks state for an API execution for emitting Metrics
type Execution struct {
	begin time.Time
}

// PlatformInstallStart creates an Execution instance and starts the timer
func PlatformInstallStart() Execution {
	return Execution{
		begin: time.Now().UTC(),
	}
}

// PlatformStackInstallationMetric is used to log duration and success/failure
func (e *Execution) PlatformStackInstallationMetric(platform *dbaasv1beta1.DBaaSPlatform, version string) {
	duration := time.Since(e.begin)
	for _, cond := range platform.Status.Conditions {
		if cond.Type == dbaasv1beta1.DBaaSPlatformReadyType {
			lastTransitionTime := cond.LastTransitionTime
			duration = lastTransitionTime.Sub(platform.CreationTimestamp.Time)
			DBaasStackInstallationHistogram.With(prometheus.Labels{MetricLabelVersion: version, MetricLabelCreationTimestamp: platform.CreationTimestamp.String()}).Observe(duration.Seconds())
		} else {
			DBaasStackInstallationHistogram.With(prometheus.Labels{MetricLabelVersion: version, MetricLabelCreationTimestamp: platform.CreationTimestamp.String()}).Observe(duration.Seconds())
		}
	}
}

// SetPlatformStatusMetric exposes dbaas_platform_status Metric for each platform
func SetPlatformStatusMetric(platformName dbaasv1beta1.PlatformsName, status dbaasv1beta1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 {
		switch status {

		case dbaasv1beta1.ResultFailed:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(status), MetricLabelVersion: version}).Set(float64(0))
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultSuccess), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultInProgress), MetricLabelVersion: version})
		case dbaasv1beta1.ResultSuccess:
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultInProgress), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultFailed), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.With(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(status), MetricLabelVersion: version}).Set(float64(1))
		case dbaasv1beta1.ResultInProgress:
			DBaasPlatformInstallationGauge.With(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(status), MetricLabelVersion: version}).Set(float64(2))
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultSuccess), MetricLabelVersion: version})
			DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultFailed), MetricLabelVersion: version})

		}
	}
}

// CleanPlatformStatusMetric delete the dbaas_platform_status Metric for each platform
func CleanPlatformStatusMetric(platformName dbaasv1beta1.PlatformsName, status dbaasv1beta1.PlatformsInstlnStatus, version string) {
	if len(platformName) > 0 && status == dbaasv1beta1.ResultSuccess {
		DBaasPlatformInstallationGauge.Delete(prometheus.Labels{MetricLabelName: string(platformName), MetricLabelStatus: string(dbaasv1beta1.ResultSuccess), MetricLabelVersion: version})
	}
}

// SetOpenShiftInstallationInfoMetric set the Metrics for openshift info
func SetOpenShiftInstallationInfoMetric(operatorVersion string, consoleURL string, platformType string) {
	DBaasOperatorVersionInfo.With(prometheus.Labels{MetricLabelVersion: operatorVersion, MetricLabelConsoleULR: consoleURL, MetricLabelPlatformName: platformType}).Set(1)
}

// UpdateRequestsDurationHistogram Utility function to update request duration histogram
func UpdateRequestsDurationHistogram(provider string, account string, namespace string, resource string, event string, duration float64) {
	DBaaSRequestsDurationHistogram.WithLabelValues(provider, account, namespace, resource, event).Observe(duration)
}

// UpdateErrorsTotal Utility function to update errors total
func UpdateErrorsTotal(provider string, account string, namespace string, resource string, event string, errCd string) {
	if len(errCd) > 0 {
		DBaaSRequestsErrorsCounter.WithLabelValues(provider, account, namespace, resource, event, errCd).Add(1)
	}
}
