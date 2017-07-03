package handler_neo4j

import (
	"fmt"
	"reflect"
	"github.com/zpatrick/go-config"

	"github.com/qnib/qframe-types"
)

const (
	version   = "0.0.0"
	pluginTyp = "handler"
	pluginPkg = "neo4j"
)

type Plugin struct {
	qtypes.Plugin
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
	bg := p.QChan.Data.Join()
	//inputs := p.GetInputs()
	for {
		select {
		case val := <-bg.Read:
			switch val.(type) {
			default:
				p.Log("info" , fmt.Sprintf("Got %s: %v", reflect.TypeOf(val), val))
			}
		}
	}
}
