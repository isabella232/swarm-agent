package config

import (
	"github.com/codegangsta/cli"
	"github.com/rancher/go-rancher/client"
)

type Config struct {
	CattleURL       string
	CattleAccessKey string
	CattleSecretKey string
	WorkerCount     int
	HealthCheckPort int
	TempDir         string
}

func Conf(context *cli.Context) Config {
	config := Config{
		CattleURL:       context.String("cattle-url"),
		CattleAccessKey: context.String("cattle-access-key"),
		CattleSecretKey: context.String("cattle-secret-key"),
		WorkerCount:     context.Int("worker-count"),
		HealthCheckPort: context.Int("health-check-port"),
		TempDir:         context.String("temp-dir"),
	}

	return config
}

func GetRancherClient(conf Config) (*client.RancherClient, error) {
	return client.NewRancherClient(&client.ClientOpts{
		Url:       conf.CattleURL,
		AccessKey: conf.CattleAccessKey,
		SecretKey: conf.CattleSecretKey,
	})
}
