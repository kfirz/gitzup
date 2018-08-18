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

	// Locate the subscription, fail if missing
	subscription := client.Subscription(config.SubscriptionName)
	log.Printf("Checking subscription '%s' exists...\n", subscription)
	exists, err := subscription.Exists(ctx)
	if err != nil {
		log.Fatalf("Failed checking if subscription exists: %v", err)
	} else if exists == false {
		log.Fatalln("Subscription could not be found!")
	}

	// Start receiving messages (in separate goroutines)
	log.Printf("Subscribing to: %s", subscription)
	err = subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("Message received: %v", msg)
		msg.Ack()
	})
	if err != nil {
		log.Fatalf("Failed to subscribe to '%s': %v", subscription, err)
	}
}
