package eventhandlers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	revents "github.com/rancher/go-machine-service/events"
	"github.com/rancher/go-rancher/client"
	util "github.com/rancher/swarm-agent/rancherevents/util"
)

type CreateHandler struct {
	baseDir string
}

func NewComposeHandler(baseDir string) *CreateHandler {
	return &CreateHandler{
		baseDir: baseDir,
	}
}

func (h *CreateHandler) Handler(event *revents.Event, cli *client.RancherClient) error {
	err := h.execute(event, cli)
	resp := util.NewReply(event)
	if err != nil {
		resp.TransitioningMessage = err.Error()
		resp.Transitioning = "error"
	}
	return util.PublishReply(resp, cli)
}

func (h *CreateHandler) execute(event *revents.Event, cli *client.RancherClient) error {
	logrus.Infof("Received event: Name: %s, Event Id: %s, Resource Id: %s", event.Name, event.ID, event.ResourceID)

	log := logrus.WithField("eventId", event.ID)

	log.Info(event)

	name := util.GetString(event.Data, "environment", "name")
	templates := util.GetStringMap(event.Data, "environment", "data", "fields", "templates")
	environment := util.GetStringMap(event.Data, "environment", "data", "fields", "environment")
	if len(templates) == 0 {
		log.Info("No templates found, returning")
		return nil
	}

	rootDir, err := h.createFiles(templates)
	if err != nil {
		return err
	}
	defer os.RemoveAll(rootDir)

	environment["DOCKER_HOST"] = os.Getenv("DOCKER_HOST")
	env := []string{}
	for k, v := range environment {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	args := []string{"-p", name, "up", "-d"}
	log.Infof("Running docker-compose %v in %s", args, rootDir)
	reader, writer := io.Pipe()
	cmd := exec.Command("docker-compose", args...)
	cmd.Dir = rootDir
	cmd.Stdout = writer
	cmd.Stderr = writer
	cmd.Env = env
	if err := cmd.Start(); err != nil {
		reader.Close()
		writer.Close()
		return err
	}

	buffer := bytes.Buffer{}
	scanner := bufio.NewScanner(reader)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		for scanner.Scan() {
			line := scanner.Text()
			buffer.WriteString(line)
			buffer.WriteString("\n")

			log.Infof("output: %s", line)

			progress := util.NewReply(event)
			progress.TransitioningMessage = line
			progress.Transitioning = "yes"
			if err := util.PublishReply(progress, cli); err != nil {
				break
			}
		}
	}()

	err = cmd.Wait()
	reader.Close()
	writer.Close()
	wg.Wait()

	if err != nil {
		err = fmt.Errorf("%v: %s", err, buffer.String())
	}

	return err
}

func (h *CreateHandler) createFiles(templates map[string]string) (string, error) {
	tempDir, err := ioutil.TempDir(h.baseDir, "docker-compose")
	if err != nil {
		return "", err
	}

	tempDir, err = filepath.Abs(tempDir)
	if err != nil {
		return "", err
	}

	if len(templates) == 1 {
		composeFile := ""
		for _, value := range templates {
			composeFile = value
		}
		templates = map[string]string{
			"docker-compose.yaml": composeFile,
		}
	}

	for name, value := range templates {
		content := fmt.Sprintf("%v", value)
		dest := path.Join(tempDir, name)
		if !strings.HasPrefix(dest, tempDir) {
			return "", fmt.Errorf("Invalid template name: %s", name)
		}

		err := ioutil.WriteFile(dest, []byte(content), 0644)
		if err != nil {
			return "", err
		}
	}

	return tempDir, nil
}
