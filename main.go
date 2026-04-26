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
		fmt.Fprintf(os.Stderr, "cannot specify both --namespace and --all-namespaces")
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
	fmt.Printf("%-10s %-12s %-20s %-12s %s\n", "TIME", "EVENT", "POD", "NAMESPACE", "PHASE")
	for event := range events.ResultChan() {
		eventType := string(event.Type)
		pod, ok := event.Object.(*v1.Pod)
		if !ok {
			fmt.Fprintf(os.Stderr, "unexpected object type: %T\n", event.Object)
			continue
		}
		coloredEventType := colorEventType(eventType)
		fmt.Printf("%-10s %s %-20s %-12s %s\n", time.Now().Format(time.TimeOnly), coloredEventType, pod.Name, pod.Namespace, pod.Status.Phase)
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
		return color.New(color.FgHiRed).Sprint(padded)
	default:
		return padded
	}
}
