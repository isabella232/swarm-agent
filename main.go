package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"

	"github.com/rancher/swarm-agent/config"
	"github.com/rancher/swarm-agent/healthcheck"
	"github.com/rancher/swarm-agent/rancherevents"
	"github.com/rancher/swarm-agent/server"
)

func main() {
	app := cli.NewApp()
	app.Name = "docker-agent"
	app.Usage = "Start the Rancher Docker agent"
	app.Action = launch

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "cattle-url",
			Usage:  "URL for cattle API",
			EnvVar: "CATTLE_URL",
		},
		cli.StringFlag{
			Name:   "cattle-access-key",
			Usage:  "Cattle API Access Key",
			EnvVar: "CATTLE_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "cattle-secret-key",
			Usage:  "Cattle API Secret Key",
			EnvVar: "CATTLE_SECRET_KEY",
		},
		cli.StringFlag{
			Name:  "temp-dir",
			Usage: "Temporary directory for content. Defaults to system temp dir",
		},
		cli.IntFlag{
			Name:   "worker-count",
			Value:  50,
			Usage:  "Number of workers for handling events",
			EnvVar: "WORKER_COUNT",
		},
		cli.IntFlag{
			Name:   "health-check-port",
			Value:  10240,
			Usage:  "Port to configure an HTTP health check listener on",
			EnvVar: "HEALTH_CHECK_PORT",
		},
	}

	app.Commands = []cli.Command{
		cli.Command{
			Name:   "server",
			Action: runServer,
		},
	}

	app.Run(os.Args)
}

func runServer(c *cli.Context) {
	f, err := ioutil.TempFile("", "swarm")
	if err != nil {
		log.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	log.Infof("Using temp file %s", f.Name())

	if len(c.Args()) == 0 {
		log.Errorf("Please pass as an arguement of the command you want to run")
		os.Exit(1)
	}

	go server.Watch(f.Name(),
		c.GlobalString("cattle-access-key"),
		c.GlobalString("cattle-secret-key"),
		c.GlobalString("cattle-url"))

	prog := c.Args()[0]
	args := append(c.Args()[1:], fmt.Sprintf("file://%s", f.Name()))

	log.Infof("Running %s %v", prog, args)
	cmd := exec.Command(prog, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		log.Errorf("Terminating: %v", err)
	}
}

func launch(c *cli.Context) {
	conf := config.Conf(c)
	resultChan := make(chan error)

	go func(rc chan error) {
		err := rancherevents.ConnectToEventStream(conf)
		log.Errorf("Rancher stream listener exited with error: %s", err)
		rc <- err
	}(resultChan)

	go func(rc chan error) {
		err := healthcheck.StartHealthCheck(conf.HealthCheckPort)
		log.Errorf("Rancher healthcheck exited with error: %s", err)
		rc <- err
	}(resultChan)

	<-resultChan
	log.Info("Exiting.")
}
