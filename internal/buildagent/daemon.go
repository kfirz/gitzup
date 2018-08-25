package buildagent

import (
	"cloud.google.com/go/pubsub"
	"golang.org/x/net/context"
	"log"
)

type Config struct {
	GcpProject       string
	SubscriptionName string
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
	err = subscription.Receive(ctx, func(_ context.Context, msg *pubsub.Message) { handleMessage(msg) })
	if err != nil {
		log.Fatalf("Failed to subscribe to '%s': %v", subscription, err)
	}
}

func handleMessage(msg *pubsub.Message) {
	defer func() {
		err := recover()
		if err != nil {
			// TODO: re-publish this message to the errors topic
			switch t := err.(type) {
			case error:
				log.Printf("Fatal error processing message '%s': %s\n", msg.ID, t.Error())
			default:
				log.Printf("Fatal error processing message: %#v\n", t)
			}
		}
	}()

	msg.Ack()

	request, err := NewBuildRequest(msg.ID, msg.Data)
	if err != nil {
		panic(err)
	}

	request.Apply()
}
