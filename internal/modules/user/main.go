package main

import (
	"net/http"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/connect"
	"github.com/sirupsen/logrus"
)

func main() {
	// Create a Consul API client
	client, _ := api.NewClient(api.DefaultConfig())
	agent := client.Agent()
	err := agent.ServiceRegister(&api.AgentServiceRegistration{
		Kind:              "microservice",
		ID:                "id",
		Name:              "user",
		Tags:              []string{"microservice", "user"},
		Port:              8080,
		Address:           "",
		TaggedAddresses:   nil,
		EnableTagOverride: false,
		Meta:              nil,
		Weights:           nil,
		Check:             nil,
		Checks: api.AgentServiceChecks{
			&api.AgentServiceCheck{
				CheckID:                        "user_grpc",
				Name:                           "check user by grpc",
				Args:                           nil,
				DockerContainerID:              "",
				Shell:                          "",
				Interval:                       (time.Second * 5).String(),
				Timeout:                        time.Second.String(),
				TTL:                            "",
				HTTP:                           "",
				Header:                         nil,
				Method:                         "",
				Body:                           "",
				TCP:                            "",
				Status:                         "",
				Notes:                          "",
				TLSSkipVerify:                  false,
				GRPC:                           "",
				GRPCUseTLS:                     false,
				AliasNode:                      "",
				AliasService:                   "",
				SuccessBeforePassing:           0,
				FailuresBeforeCritical:         0,
				DeregisterCriticalServiceAfter: "",
			},
		},
		Proxy:     nil,
		Connect:   nil,
		Namespace: "",
	})
	if err != nil {
		logrus.Fatal(err)
	}
	defer agent.ServiceDeregister("id")
	//session.Create()
	//defer session.Destroy()

	// Create an instance representing this service. "my-service" is the
	// name of _this_ service. The service should be cleaned up via Close.
	svc, _ := connect.NewService("my-service", client)
	defer svc.Close()

	mux := http.NewServeMux()
	mux.Handle("/health", http.HandlerFunc(health))
	// Creating an HTTP server that serves via Connect
	server := &http.Server{
		Addr:      ":8080",
		TLSConfig: svc.ServerTLSConfig(),
		Handler:   mux,
		// ... other standard fields
	}

	go func() {
		// Serve!
		err := server.ListenAndServe()
		if err != nil {
			logrus.Warn(err)
		}
	}()

	<-svc.ReadyWait()
	logrus.Info("jopa")
	logrus.Info(svc.Ready())
	time.Sleep(time.Hour)
}

func health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	return
}
