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
				p.handleMessage(msg)
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
