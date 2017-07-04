package handler_neo4j

import (
	"fmt"
	"reflect"
	"github.com/zpatrick/go-config"

	"github.com/qnib/qframe-types"
	"github.com/qframe/types/inventory"
)


const (
	version   = "0.1.0"
	pluginTyp = "handler"
	pluginPkg = "neo4j"
)

type Plugin struct {
	qtypes.Plugin
	NeoConn Neo4jConnector
}

func New(qChan qtypes.QChan, cfg *config.Config, name string) (Plugin, error) {
	p := Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
	}
	p.Version = version
	p.Name = name
	return p, nil
}

// Run fetches everything from the Data channel and flushes it to stdout
func (p *Plugin) Run() {
	p.Log("info", fmt.Sprintf("Start handler v%s", p.Version))
	var err error
	p.NeoConn, err = NewNeo4jConnector(p.Cfg)
	if err != nil {
		return
	}
	defer p.NeoConn.Close()
	bg := p.QChan.Data.Join()
	for {
		select {
		case val := <-bg.Read:
			switch val.(type) {
			case qtypes.Message:
				qm := val.(qtypes.Message)
				if p.StopProcessingMessage(qm, false) {
					continue
				}
			case qtypes_inventory.Base:
				inv := val.(qtypes_inventory.Base)
				if inv.StopProcessing(p.Plugin, false) {
					continue
				}
				p.HandleInventoryBase(inv)
			default:
				p.Log("info" , fmt.Sprintf("Got %s: %v", reflect.TypeOf(val), val))
			}
		}
	}
}


func (p *Plugin) HandleInventoryBase(ib qtypes_inventory.Base) {
	err := p.NeoConn.ConnectBase(ib)
	if err != nil {
		p.Log("error", err.Error())
	}
	msg := fmt.Sprintf("Got inventory: Subject '%s' '%s' object '%s' (Tags:%v)", ib.Subject, ib.Action, ib.Object, ib.Tags)
	p.Log("info", msg)

}