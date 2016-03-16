package server

import (
	"fmt"
	"sort"

	"github.com/Sirupsen/logrus"
	"github.com/rancher/go-rancher-metadata/metadata"
	"github.com/rancher/go-rancher/client"
	dockerapiproxy "github.com/rancher/rancher-docker-api-proxy"
)

const (
	base = 3000
	max  = 40000
)

type Proxy struct {
	client  *client.RancherClient
	ports   map[string]int
	proxies map[int]*dockerapiproxy.Proxy
}

func NewProxy(c *client.RancherClient) *Proxy {
	return &Proxy{
		client:  c,
		ports:   map[string]int{},
		proxies: map[int]*dockerapiproxy.Proxy{},
	}
}

func (p *Proxy) AddHosts(hosts []metadata.Host) []string {
	result := []string{}
	newHosts := map[string]bool{}

	for _, host := range hosts {
		newHosts[host.UUID] = true
		if _, ok := p.ports[host.UUID]; ok {
			continue
		}
		port := p.next()
		proxy := dockerapiproxy.NewProxy(p.client, host.UUID, fmt.Sprintf("tcp://localhost:%d", port))

		p.ports[host.UUID] = port
		p.proxies[port] = proxy

		go func(proxy *dockerapiproxy.Proxy, uuid string) {
			if err := proxy.ListenAndServe(); err != nil {
				logrus.Errorf("ListenAndServe: %s :%v", uuid, err)
			}
		}(proxy, host.UUID)
	}

	for host, port := range p.ports {
		if newHosts[host] {
			continue
		}

		p.proxies[port].Close()
		delete(p.proxies, port)
		delete(p.ports, host)
	}

	for _, host := range hosts {
		result = append(result, fmt.Sprintf("localhost:%d", p.ports[host.UUID]))
	}

	sort.Sort(sort.StringSlice(result))
	return result
}

func (p *Proxy) next() int {
	for i := base; i < max; i++ {
		if _, ok := p.proxies[i]; !ok {
			return i
		}
	}
	panic("No free ports")
}
