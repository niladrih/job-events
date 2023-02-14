package main

import (
	"context"
	"fmt"
	"os"

	"github.com/openebs/dynamic-localpv-provisioner/pkg/kubernetes/api/core/v1/event"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	cmd := NewUpgradectlCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func NewUpgradectlCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:          "upgradectl",
		Short:        "Demo tool for Mayastor Upgrade",
		Long:         "It is what it is...",
		SilenceUsage: true,
	}
	cmd.AddCommand(upgradeStatus())

}

func upgradeStatus() *cobra.Command {
	const (
		jobName      = "mayastor-job-upgrade-v2-1-0"
		jobNamespace = "mayastor"
	)

	cmd := &cobra.Command{
		Use: "upgrade-status",
		Run: func(cmd *cobra.Command, _ []string) {
			cfg, err := clientcmd.BuildConfigFromFlags("", "/root/.kube/config")
			clientset, err := kubernetes.NewForConfig(cfg)

			fieldSelector := "involvedObject.kind=Job" + "," +
				"involvedObject.name=" + jobName
			k8sApiEventList, err := clientset.CoreV1().Events(jobNamespace).List(context.TODO(),
				metav1.ListOptions{FieldSelector: fieldSelector})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to get Events for Job %s in %s namespace", jobName, jobNamespace)
				os.Exit(1)
			}
			if len(k8sApiEventList.Items) == 0 {
				fmt.Fprintf(os.Stderr, "Not enough Events for Job %s in %s namespace", jobName, jobNamespace)
				os.Exit(1)
			}
			eventList := event.ListBuilderFromAPIList(k8sApiEventList).List().LatestFirstSort()
			fmt.Printf("Reason: %s\nMessage: %s\n", eventList.Items[0].Object.Reason, eventList.Items[0].Object.Message)
		},
	}
}
