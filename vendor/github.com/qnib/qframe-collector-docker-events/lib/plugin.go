package qframe_collector_docker_events

import (
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/zpatrick/go-config"
	"golang.org/x/net/context"

	"github.com/qnib/qframe-inventory/lib"
	"github.com/qnib/qframe-types"
	"strings"
	"time"
)

const (
	version   = "0.2.4"
	pluginTyp = qtypes.COLLECTOR
	pluginPkg = "docker-events"
	dockerAPI = "v1.29"
)

type Plugin struct {
	qtypes.Plugin
	engCli *client.Client
	info types.Info
}

func New(qChan qtypes.QChan, cfg *config.Config, name string) (Plugin, error) {
	var err error
	p := Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
	}
	return p, err
}

func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start docker-events collector v%s", p.Version))
	ctx := context.Background()
	dockerHost := p.CfgStringOr("docker-host", "unix:///var/run/docker.sock")
	// Filter start/stop event of a container
	engineCli, err := client.NewClient(dockerHost, dockerAPI, nil, nil)
	if err != nil {
		p.Log("error", fmt.Sprintf("Could not connect docker/docker/client to '%s': %v", dockerHost, err))
		return
	}
	p.info, err = engineCli.Info(context.Background())
	if err != nil {
		p.Log("error", fmt.Sprintf("Error during Info(): %v >err> %s", p.info, err))
		return
	} else {
		p.Log("info", fmt.Sprintf("Connected to '%s' / v'%s'", p.info.Name, p.info.ServerVersion))
	}
	// Inventory Init
	inv := qframe_inventory.NewInventory()
	// Fire events for already started containers
	cnts, _ := engineCli.ContainerList(ctx, types.ContainerListOptions{})
	for _, cnt := range cnts {
		cJson, err := engineCli.ContainerInspect(ctx, cnt.ID)
		if err != nil {
			continue
		}
		p.Log("debug", fmt.Sprintf("Already running container %s: SetItem(%s)", cJson.Name, cJson.ID))
		inv.SetItem(cnt.ID, cJson)
	}
	msgs, errs := engineCli.Events(context.Background(), types.EventsOptions{})
	for {
		select {
		case dMsg := <-msgs:
			base := qtypes.NewTimedBase(p.Name, time.Unix(dMsg.Time, 0))
			if dMsg.Type == "container" {
				data := map[string]string{"args": ""}
				if strings.HasPrefix(dMsg.Action, "exec_") {
					exec := strings.Split(dMsg.Action, ":")
					dMsg.Action = exec[0]
					data["args"] = strings.Join(exec[1:], " ")
				}
				data["action"] = dMsg.Action
				cnt, err := inv.GetItem(dMsg.Actor.ID)
				if err != nil {
					switch dMsg.Action {
					case "die", "destroy":
						p.Log("debug", fmt.Sprintf("Container %s just '%s' without having an entry in the Inventory", dMsg.Actor.ID, dMsg.Action))
						continue
					case "create", "attach", "commit", "resize":
						continue
					case "start":
						cnt, err := engineCli.ContainerInspect(ctx, dMsg.Actor.ID)
						if err != nil {
							p.Log("error", fmt.Sprintf("Could not inspect '%s'", dMsg.Actor.ID))
							continue
						}
						inv.SetItem(dMsg.Actor.ID, cnt)
						ce := qtypes.NewContainerEvent(base, cnt, dMsg, p.info)
						for k, v := range data {
							ce.Data[k] = v
						}
						ce.Message = fmt.Sprintf("%s: %s.%s %v", dMsg.Actor.Attributes["name"], dMsg.Type, dMsg.Action, data)
						p.Log("debug", fmt.Sprintf("Just started container %s: SetItem(%s)", cnt.Name, cnt.ID))
						p.QChan.Data.Send(ce)
						continue
					}
				}

				p.Log("debug", fmt.Sprintf("Container '%s' was found in the inventory...", dMsg.Actor.Attributes["name"]))
				if err != nil {
					msg := fmt.Sprintf("Could not find container '%s' in invntory while it is doing '%s.%s'", dMsg.Actor.ID, dMsg.Type, dMsg.Action)
					p.Log("error", msg)
					continue
				}
				ce := qtypes.NewContainerEvent(base, cnt, dMsg, p.info)
				for k, v := range data {
					ce.Data[k] = v
				}
				ce.Message = fmt.Sprintf("%s: %s.%s %v", dMsg.Actor.Attributes["name"], dMsg.Type, dMsg.Action, data)
				p.Log("debug", fmt.Sprintf("Just started container %s: SetItem(%s)", cnt.Name, cnt.ID))
				p.QChan.Data.Send(ce)
				continue
			}
		case dErr := <-errs:
			if dErr != nil {
				p.Log("error", dErr.Error())
			}
		}
	}
}
