package server

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/go-rancher/client"
)

const (
	metadataUrl = "http://169.254.169.250/2015-12-19"
)

func Watch(file, accessKey, secretKey, url string) error {
	logrus.Infof("Watching for changes %s %s", accessKey, url)
	m, err := metadata.NewClientAndWait(metadataUrl)
	if err != nil {
		return err
	}

	client, err := client.NewRancherClient(&client.ClientOpts{
		AccessKey: accessKey,
		SecretKey: secretKey,
		Url:       url,
	})
	if err != nil {
		return err
	}

	proxy := NewProxy(client)

	for {
		time.Sleep(2 * time.Second)

		hosts, err := m.GetHosts()
		if err != nil {
			logrus.Errorf("Error gettings hosts: %v", err)
			continue
		}

		fileContent := proxy.AddHosts(hosts)
		if err := WriteFile(file, strings.Join(fileContent, "\n")); err != nil {
			logrus.Errorf("Failed to write [%s] to file %s: %v", strings.Join(fileContent, "\n"), file, err)
		}
	}
}

func WriteFile(file, content string) error {
	err := ioutil.WriteFile(file+".tmp", []byte(content), 0644)
	if err != nil {
		return err
	}

	return os.Rename(file+".tmp", file)
}
