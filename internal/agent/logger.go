package agent

import (
	"context"
	log "github.com/sirupsen/logrus"
)

func From(ctx context.Context) *log.Entry {
	var logger = log.NewEntry(log.StandardLogger())
	if requestId, ok := ctx.Value("request").(string); ok {
		logger = logger.WithField("request", requestId)
	}
	if resourceName, ok := ctx.Value("resource").(string); ok {
		logger = logger.WithField("resource", resourceName)
	}
	if containerName, ok := ctx.Value("container").(string); ok {
		logger = logger.WithField("container", containerName)
	}
	return logger
}
