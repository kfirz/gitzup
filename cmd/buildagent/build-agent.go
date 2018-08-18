package main

import (
	"flag"
	"github.com/kfirz/gitzup/internal/buildagent"
	"log"
)

func main() {

	// Parse command-line
	var project, topic, subscription string
	flag.StringVar(&project, "project", "", "Google Cloud project ID (required)")
	flag.StringVar(&topic, "topic", "", "Google Cloud Pub/Sub topic (required)")
	flag.StringVar(&subscription, "subscription", "", "Google Cloud Pub/Sub subscription (required)")
	flag.Parse()

	// Validate flags
	if project == "" {
		log.Fatalln("GCP project is required")
		flag.Usage()
	}
	if topic == "" {
		log.Fatalln("GCP Pub/Sub topic is required")
		flag.Usage()
	}
	if subscription == "" {
		log.Fatalln("GCP Pub/Sub subscription is required")
		flag.Usage()
	}

	// Start daemon
	buildagent.Daemon(buildagent.Config{GcpProject: project, TopicName: topic, SubscriptionName: subscription})

}
