package collector_file

import (
	"log"
	"os"

	"github.com/hpcloud/tail"
	"github.com/qnib/qframe-types"
	"github.com/zpatrick/go-config"
)

const (
	version = "0.1.1"
	pluginTyp = "collector"
	pluginPkg = "file"
)

type Plugin struct {
	qtypes.Plugin
	Name string
}

func New(qChan qtypes.QChan, cfg *config.Config, name string) (Plugin, error) {
	return Plugin{
		Plugin: qtypes.NewNamedPlugin(qChan, cfg, pluginTyp, pluginPkg, name, version),
	}, nil
}

func (p *Plugin) Run() {
	log.Printf("[II] Start collector v%s", version)
	fPath, err := p.CfgString("path")
	if err != nil {
		log.Println("[EE] No file path for collector.file.path set")
		return
	}
	create := p.CfgBoolOr("collector.file.create", false)
	if _, err := os.Stat(fPath); os.IsNotExist(err) && create {
		log.Printf("[DD] Create file: %s", fPath)
		f, _ := os.Create(fPath)
		f.Close()
	}
	fileReopen, err := p.Cfg.BoolOr("collector.file.reopen", true)
	t, err := tail.TailFile(fPath, tail.Config{Follow: true, ReOpen: fileReopen})
	if err != nil {
		log.Printf("[WW] File collector failed to open %s: %s", fPath, err)
	}
	b := qtypes.NewBase(p.Name)
	for line := range t.Lines {
		qm := qtypes.NewMessage(b, fPath, qtypes.MsgFile, line.Text)
		p.QChan.Data.Send(qm)
	}
}
