package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

// EnabledService is the type for HCA enable services from Swarm
type EnabledService struct {
	ID          string
	Name        string
	Spec        swarm.ServiceSpec
	Version     swarm.Version
	Replicas    *uint64
	MinReplicas uint64
	MaxReplicas uint64
}

func main() {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	if err != nil {
		panic(err)
	}

	services, err := cli.ServiceList(ctx, types.ServiceListOptions{})
	if err != nil {
		panic(err)
	}

	var enabledServices []EnabledService

	if len(services) > 0 {
		for _, service := range services {
			if labels := service.Spec.Labels; labels["hca.enable"] == "true" {
				minRepl, _ := strconv.ParseUint(labels["hca.min-containers"], 10, 64)
				maxRepl, _ := strconv.ParseUint(labels["hca.max-containers"], 10, 64)

				enabledServices = append(enabledServices, EnabledService{
					ID:          service.ID,
					Name:        service.Spec.Name,
					Spec:        service.Spec,
					Version:     service.Version,
					Replicas:    service.Spec.Mode.Replicated.Replicas,
					MinReplicas: minRepl,
					MaxReplicas: maxRepl,
				})
			}
		}
	} else {
		fmt.Printf("No services found")
		return
	}

	for _, service := range enabledServices {
		*service.Replicas = service.MinReplicas
		response, err := cli.ServiceUpdate(
			ctx,
			service.ID,
			service.Version,
			service.Spec,
			types.ServiceUpdateOptions{},
		)
		if err != nil {
			fmt.Println(err)
			return
		}

		fmt.Println(response)

		fmt.Printf(
			"Service %v scaled to %v\n",
			service.Name,
			*service.Replicas,
		)
	}

}
