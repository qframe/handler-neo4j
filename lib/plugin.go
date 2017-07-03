package handler_neo4j

import (
	"fmt"
	"reflect"
	"github.com/zpatrick/go-config"
	bolt "github.com/johnnadratowski/golang-neo4j-bolt-driver"

	"github.com/qnib/qframe-types"
	"github.com/qframe/types/inventory"
)


const (
	version   = "0.0.0"
	pluginTyp = "handler"
	pluginPkg = "neo4j"
)

type Plugin struct {
	qtypes.Plugin
	conn bolt.Conn
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
	p.conn, err = p.Connect()
	if err != nil {
		return
	}
	defer p.conn.Close()
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

func (p *Plugin) Connect() (bolt.Conn, error) {
	host := p.CfgStringOr("host", "localhost")
	port := p.CfgStringOr("port", "7687")
	addr := fmt.Sprintf("bolt://%s:%s", host, port)
	conn, err := bolt.NewDriver().OpenNeo(addr)
	if err != nil {
		p.Log("error", fmt.Sprintf("Not able to connect to '%s': %s", addr, err.Error()))
	}
	return conn, err
}
