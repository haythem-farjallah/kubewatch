package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	var nameSpace string
	flag.StringVar(&nameSpace, "n", "default", "namespace (shorthand)")
	flag.StringVar(&nameSpace, "namespace", "default", "namespace")
	var allNamespace bool
	flag.BoolVar(&allNamespace, "all-namespaces", false, "list pods across all namespace")
	flag.BoolVar(&allNamespace, "A", false, "list pods across all namespace (shorthand)")
	var selector string
	flag.StringVar(&selector, "selector", "", "filter pods by label selector")
	flag.StringVar(&selector, "l", "", "filter pods by label selector (shorthand)")
	flag.Parse()
	namespaceWasSet := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "n" || f.Name == "namespace" {
			namespaceWasSet = true
		}
	})

	if allNamespace && namespaceWasSet {
		fmt.Fprintln(os.Stderr, "cannot specify both --namespace and --all-namespaces")
		os.Exit(1)
	} else if allNamespace {
		nameSpace = metav1.NamespaceAll
	}
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load kubeconfig : %v\n", err)
		os.Exit(1)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create clientset : %v\n", err)
		os.Exit(1)
	}
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to reach cluster : %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✅ connected to cluster \n")
	fmt.Printf("	Server version: %s\n", version.GitVersion)
	fmt.Printf("	Platform:       %s\n", version.Platform)
	ctx := context.Background()
	events, err := clientset.CoreV1().Pods(nameSpace).Watch(ctx, metav1.ListOptions{LabelSelector: selector})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to reach cluster : %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("%-10s %-12s %-20s %-12s %-12s %-12s %-12s %s\n", "TIME", "EVENT", "POD", "NAMESPACE", "READY", "STATUS", "RESTARTS", "NODE")
	for event := range events.ResultChan() {
		eventType := string(event.Type)
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			fmt.Fprintf(os.Stderr, "unexpected object type: %T\n", event.Object)
			continue
		}
		coloredEventType := colorEventType(eventType)
		fmt.Printf("%-10s %-12s %-20s %-12s %-12s %-12s %-12d %-12s \n",
			time.Now().Format(time.TimeOnly),
			coloredEventType,
			pod.Name,
			pod.Namespace,
			podReady(pod),
			pod.Status.Phase,
			podRestarts(pod),
			podNode(pod))
	}
}
func colorEventType(eventType string) string {
	padded := fmt.Sprintf("%-12s", eventType)
	switch eventType {
	case "ADDED":
		return color.New(color.FgGreen).Sprint(padded)
	case "MODIFIED":
		return color.New(color.FgYellow).Sprint(padded)
	case "DELETED":
		return color.New(color.FgRed).Sprint(padded)
	case "ERROR":
		return color.New(color.FgHiRed, color.Bold).Sprint(padded)
	default:
		return padded
	}
}
func podReady(pod *v1.Pod) string {
	readyCount := 0
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			readyCount++
		}
	}
	return fmt.Sprintf("%d/%d", readyCount, len(pod.Spec.Containers))
}

func podRestarts(pod *v1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}
func podNode(pod *v1.Pod) string {
	if pod.Spec.NodeName == "" {
		return "<none>"
	}
	return pod.Spec.NodeName
}
