package main

import (
	"context"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
)

const (
	defaultJobName         = "mayastor-job-upgrade-v2-1-0"
	defaultJobNamespace    = "mayastor"
	componentName          = "mayastor-upgrade-job"
	eventLoopSleepDuration = "5s"
)

func main() {
	klog.InitFlags()
	defer klog.Flush()

	ctx := context.Background()
	var jobName string
	var jobNamespace string

	jobName, ok := os.LookupEnv("JOB_NAME")
	if !ok {
		jobName = defaultJobName
	}

	jobNamespace, ok = os.LookupEnv("JOB_NAMESPACE")
	if !ok {
		jobNamespace = defaultJobNamespace
	}

	cfg, _ := rest.InClusterConfig()
	clientset, _ := kubernetes.NewForConfig(cfg)

	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartLogging(klog.Infof)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: clientset.CoreV1().Events(jobNamespace)})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: componentName})

	job, _ := clientset.BatchV1().Jobs(jobNamespace).Get(ctx, jobName, metav1.GetOptions{})

	for {
		startEventing(recorder, job)
		klog.Infof("Looping over the events again in %s", eventLoopSleepDuration)
		sleepDuration, _ := time.ParseDuration(eventLoopSleepDuration)
		time.Sleep(sleepDuration)
	}
}

func startEventing(recorder record.EventRecorder, obj runtime.Object) {
	const (
		upgradingControlPlaneState = "UpgradingControlPlane"
		upgradingDataPlaneState    = "UpgradingDataPlane"
		//errorState                 = "ErrorUpgrading"
		completedState = "SuccessfulUpgrade"

		currentSemver = "2.0.0"
		targetSemver  = "2.1.0"

		durationBetweenEvents = "15s"
	)
	sleepDuration, _ := time.ParseDuration(durationBetweenEvents)

	recorder.Eventf(obj, corev1.EventTypeNormal, upgradingControlPlaneState,
		"Upgrading helm release for Mayastor from v%s to v%s", currentSemver, targetSemver)
	klog.Infof("Sleeping for %s...", durationBetweenEvents)
	time.Sleep(sleepDuration)

	recorder.Eventf(obj, corev1.EventTypeNormal, upgradingDataPlaneState,
		"Upgrading Mayastor data plane from v%s to v%s", currentSemver, targetSemver)
	klog.Infof("Sleeping for %s...", durationBetweenEvents)
	time.Sleep(sleepDuration)

	recorder.Eventf(obj, corev1.EventTypeNormal, upgradingDataPlaneState,
		"Successful Upgrade from Mayastor v%s to v%s", currentSemver, targetSemver)
	klog.Info("fin")
}
