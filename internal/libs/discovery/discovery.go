// Package discovery this is wrapper for consul.
package discovery

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/consul/api"
)

type (
	// Discovery client for register and deregister and search service.
	Discovery struct {
		consul *api.Client
	}
)

// New build and returns new Discovery client.
func New(addr string) (*Discovery, error) {
	cfg := api.DefaultConfig()
	cfg.Address = addr

	consul, err := api.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create new instance discovery client: %w", err)
	}

	return &Discovery{consul: consul}, nil
}

// Register service in consul.
func (c *Discovery) Register(serviceName, id string, ip net.IP, httpPort int, tags ...string) error {
	agent := c.consul.Agent()

	host := net.JoinHostPort(ip.String(), strconv.Itoa(httpPort))
	externalAPICheck := &api.AgentServiceCheck{
		CheckID:  strings.Join([]string{id, "external"}, "-"),
		Name:     fmt.Sprintf("Check external api by addr %s", host),
		Interval: (time.Second * 60).String(),
		Timeout:  (time.Second * 5).String(),
		HTTP:     fmt.Sprintf("http://%s/health", host),
		Method:   http.MethodGet,
	}

	arg := api.AgentServiceRegistration{
		Kind:    api.ServiceKindTypical,
		ID:      id,
		Name:    serviceName,
		Tags:    tags,
		Port:    httpPort,
		Address: ip.String(),
		Checks:  []*api.AgentServiceCheck{externalAPICheck},
	}

	err := agent.ServiceRegister(&arg)
	if err != nil {
		return fmt.Errorf("register service by arg: %+v err: %w", arg, err)
	}

	return nil
}

// Deregister service from consul.
func (c *Discovery) Deregister(id string) error {
	agent := c.consul.Agent()

	err := agent.ServiceDeregister(id)
	if err != nil {
		return fmt.Errorf("deregister service by id: %s err: %w", id, err)
	}

	return nil
}

// Errors.
var (
	ErrCfgNotFound = errors.New("config not found")
)

// Config gets configuration from consul and unmarshal in val.
func (c *Discovery) Config(ctx context.Context, serviceName string, val interface{}) error {
	q := &api.QueryOptions{}
	q.WithContext(ctx)
	key := fmt.Sprintf("config/%s", serviceName)
	kv, _, err := c.consul.KV().Get(key, q)
	if err != nil {
		return fmt.Errorf("get kv from discovery: %w by key: %s", err, key)
	}

	if kv == nil {
		return fmt.Errorf("%w by key %s", ErrCfgNotFound, key)
	}

	err = json.Unmarshal(kv.Value, val)
	if err != nil {
		return fmt.Errorf("unmarshal cfg: %w", err)
	}

	return nil
}

// ServiceAddr
func (c *Discovery) ServiceAddr(id string) (string, error) {
	srv, _, err := c.consul.Agent().Service(id, nil)
	if err != nil {
		return "", fmt.Errorf("get service info by id: %s err: %w", id, err)
	}

	return net.JoinHostPort(srv.Address, strconv.Itoa(srv.Port)), nil
}
