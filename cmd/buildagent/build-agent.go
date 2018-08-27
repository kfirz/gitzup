package main

import (
	"flag"
	"github.com/kfirz/gitzup/internal/buildagent"
	"log"
)

func main() {

	// Parse command-line
	var project, subscription string
	flag.StringVar(&project, "project", "gitzup", "Google Cloud project ID (required)")
	flag.StringVar(&subscription, "subscription", "agents", "Google Cloud Pub/Sub subscription (required)")
	flag.Parse()

	// Validate flags
	if project == "" {
		log.Fatalln("GCP project is required")
		flag.Usage()
	}

	if subscription == "" {
		log.Fatalln("GCP Pub/Sub subscription is required")
		flag.Usage()
	}

	// Start daemon
	buildagent.Daemon(buildagent.Config{GcpProject: project, SubscriptionName: subscription})

}
