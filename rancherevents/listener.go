package rancherevents

import (
	revents "github.com/rancher/event-subscriber/events"
	"github.com/rancher/swarm-agent/config"
	"github.com/rancher/swarm-agent/rancherevents/eventhandlers"
)

func ConnectToEventStream(conf config.Config) error {

	eventHandlers := map[string]revents.EventHandler{
		"composeProject.create": eventhandlers.NewComposeHandler(conf.TempDir).Handler,
		"ping":                  eventhandlers.NewPingHandler().Handler,
	}

	router, err := revents.NewEventRouter("", 0, conf.CattleURL, conf.CattleAccessKey, conf.CattleSecretKey, nil, eventHandlers, "", conf.WorkerCount, revents.DefaultPingConfig)
	if err != nil {
		return err
	}
	err = router.StartWithoutCreate(nil)
	return err
}
