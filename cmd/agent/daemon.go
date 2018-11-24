package main

import (
	"cloud.google.com/go/pubsub"
	"context"
	"github.com/go-errors/errors"
	"github.com/kfirz/gitzup/internal/agent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the Gitzup agent daemon.",
	Long:  `This command will start the Gitzup agent daemon, processing build request coming in through the GCP Pub/Sub subscription.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("GCP project ID is required")
		}
		if len(args) < 2 {
			return errors.New("GCP Pub/Sub subscription name is required")
		}
		startDaemon(args[0], args[1])
		return nil
	},
}


// Initializes the main package with global flags
func init() {
	rootCmd.AddCommand(daemonCmd)
}

func startDaemon(gcpProject string, gcpSubscriptionName string) {

	// Create context for the daemon
	ctx := context.Background()

	// Create the Pub/Sub client
	client, err := pubsub.NewClient(ctx, gcpProject)
	if err != nil {
		log.WithError(err).Fatal("Could not create Pub/Sub client")
	}
	defer func() {
		err = client.Close()
		if err != nil {
			log.WithError(err).Error("Could not close PubSub client")
		}
	}()

	// Locate the subscription, fail if missing
	subscription := client.Subscription(gcpSubscriptionName)
	exists, err := subscription.Exists(ctx)
	if err != nil {
		log.WithError(err).Fatalf("Failed verifying that subscription '%s' exists", gcpSubscriptionName)
	} else if exists == false {
		log.WithError(err).Fatalf("Could not find subscription '%s'", subscription)
	}

	// Start receiving messages (in separate goroutines)
	log.Infof("Subscribing to: %s", subscription)
	err = subscription.Receive(ctx, func(_ context.Context, msg *pubsub.Message) { handleMessage(msg) })
	if err != nil {
		log.WithError(err).Fatalf("Could not subscribe to '%s'", subscription)
	}
}

func handleMessage(msg *pubsub.Message) {
	defer func() {
		err := recover()
		if err != nil {
			// TODO: re-publish this message to the errors topic?
			switch t := err.(type) {
			case error:
				// TODO: print with stacktrace
				log.WithError(t).Errorf("Failed processing message '%s'", msg.ID)
			default:
				// TODO: print with stacktrace
				log.Errorf("Failed processing message '%s': %#v", msg.ID, t)
			}
		}
	}()

	msg.Ack()

	request, err := agent.New(msg.ID, workspacePath, msg.Data)
	if err != nil {
		panic(err)
	}

	// TODO: timeout support should be provided as metadata on the pub/sub message
	err = request.Apply(context.WithValue(context.Background(), "request", request.Id()))
	if err != nil {
		panic(err)
	}

	// TODO: receive apply result, and send Pub/Sub message with JSON of result
}
