package main

import (
	"log"
	"github.com/zpatrick/go-config"
	"github.com/qnib/qframe-types"
	"github.com/qframe/handler-neo4j/lib"
	"github.com/qframe/collector-file/lib"
	"sync"
)


func Run(qChan qtypes.QChan, cfg *config.Config, name string) {
	p, _ := handler_neo4j.New(qChan, cfg, name)
	p.Run()
}


func main() {
	qChan := qtypes.NewQChan()
	qChan.Broadcast()
	cfgMap := map[string]string{
		"collector.file.path": "./resources/inventory.events",
	}

	cfg := config.NewConfig([]config.Provider{config.NewStatic(cfgMap)})


	n4j, err := handler_neo4j.New(qChan, cfg, "neo4j")
	if err != nil {
		log.Printf("[EE] Failed to create collector: %v", err)
		return
	}
	go n4j.Run()
	// Dummy file
	cf, err := collector_file.New(qChan, cfg, "file")
	if err != nil {
		log.Printf("[EE] Failed to create collector: %v", err)
		return
	}
	go cf.Run()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}