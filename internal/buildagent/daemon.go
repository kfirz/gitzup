package buildagent

import (
	"cloud.google.com/go/pubsub"
	"github.com/kfirz/gitzup/internal/pipeline"
	"golang.org/x/net/context"
	"log"
)

func work(pipelinePath string) {
	p, err := pipeline.ParsePipeline(pipelinePath)
	if err != nil {
		log.Fatalf("Pipeline '%s' could not be parsed: %s\n", pipelinePath, err.Error())
	}

	err = p.Build()
	if err != nil {
		log.Fatalf("Pipeline '%s' could not be built: %s\n", pipelinePath, err.Error())
	}
}

func Daemon(config Config) {

	// Create context for the daemon
	ctx := context.Background()

	// Creates the PubSub client.
	client, err := pubsub.NewClient(ctx, config.GcpProject)
	if err != nil {
		log.Fatalf("Failed to create GCP Pub/Sub client: %v", err)
	}
	defer client.Close()

	// Start receiving messages (in separate goroutines)
	subscription := client.SubscriptionInProject(config.SubscriptionName, config.GcpProject)
	err = subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Println(msg)
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to '%s': %v", subscription, err)
	}
}
