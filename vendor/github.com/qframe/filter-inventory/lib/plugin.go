package filter_inventory

import (
	"fmt"
	"reflect"
	"strings"
	"github.com/zpatrick/go-config"

	"github.com/qnib/qframe-types"
	"github.com/qframe/types/inventory"
	"github.com/qframe/types/messages"
)

const (
	version = "0.1.0"
	pluginTyp = qtypes.FILTER
	pluginPkg = "inventory"
)

type Plugin struct {
	qtypes.Plugin
}

func New(qChan qtypes.QChan, cfg *config.Config, name string) (Plugin, error) {
	return Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
	}, nil
}

func (p *Plugin) Run() {
	p.Log("notice", fmt.Sprintf("Start inventory v%s", p.Version))
	dc := p.QChan.Data.Join()
	for {
		select {
		case val := <-dc.Read:
			switch val.(type) {
			case qtypes.Message:
				msg := val.(qtypes.Message)
				p.StopProcessingMessage(msg, false)
				p.handleMessage(msg)
			case qtypes.ContainerEvent:
				ce := val.(qtypes.ContainerEvent)
				p.StopProcessingCntEvent(ce, false)
				p.handleContainerEvent(ce)
			default:
				p.Log("trace", fmt.Sprintf("Received %s: %v", reflect.TypeOf(val), val))
			}
		}
	}
}

func (p *Plugin) handleMessage(msg qtypes.Message) {
	switch msg.MessageType {
	case qtypes.MsgFile:
		p.Log("info", fmt.Sprintf("Received '%s' Message:%s", msg.Name, msg.Message))
		if strings.HasSuffix(msg.Name, ".json") {
			qb := qtypes_messages.NewBaseFromOldBase(p.Name, msg.Base)
			b, err := qtypes_inventory.NewBaseFromJson(qb, msg.Message)
			if err != nil {
				p.Log("error", fmt.Sprintf("Error parsing '%s': %s", msg.Message, err.Error()))
				return
			}
			p.QChan.SendData(b)

		}
	}
}

func (p *Plugin) handleContainerEvent(ce qtypes.ContainerEvent) {
	switch ce.Event.Action {
	case "resize":
		return
	default:
		p.Log("debug", fmt.Sprintf("Received '%s.%s' for '%s'", ce.Event.Type, ce.Event.Action, ce.Container.Name))
		b, err := qtypes_inventory.NewBaseFromContainerEvent(ce)
		if err != nil {
			p.Log("error", fmt.Sprintf("Error creating new base from ContainerEvent: %s", err.Error()))
			return
		}
		b.EnrichContainer(ce)
		p.QChan.SendData(b)
	}

}
