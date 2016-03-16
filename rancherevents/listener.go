package rancherevents

import (
	revents "github.com/rancher/go-machine-service/events"
	"github.com/rancher/swarm-agent/config"
	"github.com/rancher/swarm-agent/rancherevents/eventhandlers"
)

func ConnectToEventStream(conf config.Config) error {

	eventHandlers := map[string]revents.EventHandler{
		"composeProject.create": eventhandlers.NewComposeHandler(conf.TempDir).Handler,
		"ping":                  eventhandlers.NewPingHandler().Handler,
	}

	router, err := revents.NewEventRouter("", 0, conf.CattleURL, conf.CattleAccessKey, conf.CattleSecretKey, nil, eventHandlers, "", conf.WorkerCount)
	if err != nil {
		return err
	}
	err = router.StartWithoutCreate(nil)
	return err
}
