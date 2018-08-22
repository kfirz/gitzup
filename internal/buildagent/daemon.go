package buildagent

import (
	"cloud.google.com/go/pubsub"
	"fmt"
	"golang.org/x/net/context"
	"log"
)

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
	err = subscription.Receive(ctx, handleMessage)
	if err != nil {
		log.Fatalf("Failed to subscribe to '%s': %v", subscription, err)
	}
}

func handleMessage(_ context.Context, msg *pubsub.Message) {
	defer func() {
		err := recover()
		if err != nil {
			// TODO: re-publish this message to the errors topic
			log.Printf("Fatal error processing message: %#v\n", err)
		}
	}()

	msg.Ack()

	var request *BuildRequest
	var err error
	switch msg.Attributes["version"] {
	default:
		request, err = handleMessageV1(msg)
		if err != nil {
			panic(fmt.Sprintf("Failed processing v1 build request message '%s': %#v", string(msg.Data), err))
		}
	}

	request.Apply()
}

func handleMessageV1(msg *pubsub.Message) (request *BuildRequest, err error) {
	request, err = NewBuildRequestV1(msg.Attributes, msg.Data)
	if err != nil {
		return nil, err
		panic(fmt.Sprintf("Failed processing a v1 build request message '%s': %#v", string(msg.Data), err))
	}

	log.Printf("Received build request: %#v\n", request)
	return request, err
}
