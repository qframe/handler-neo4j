package main

import (
	"log"
	"github.com/zpatrick/go-config"
	"github.com/qnib/qframe-types"
	"github.com/qframe/handler-neo4j/lib"
	"github.com/qframe/collector-file/lib"
	"sync"
	"github.com/qframe/filter-inventory/lib"
	"github.com/qnib/qframe-collector-docker-events/lib"
)


func Run(qChan qtypes.QChan, cfg *config.Config, name string) {
	p, _ := handler_neo4j.New(qChan, cfg, name)
	p.Run()
}


func main() {
	qChan := qtypes.NewQChan()
	qChan.Broadcast()
	cfgMap := map[string]string{
		"collector.file.path": "./resources/inventory.json",
		"filter.inventory.inputs": "file",
		"handler.neo4j.inputs": "inventory",
		"log.level": "info",
	}

	cfg := config.NewConfig([]config.Provider{config.NewStatic(cfgMap)})


	n4j, err := handler_neo4j.New(qChan, cfg, "neo4j")
	if err != nil {
		log.Printf("[EE] Failed to create (%s) %s: %v", n4j.Typ, n4j.Name, err)
		return
	}
	go n4j.Run()
	// Filter to parse string to BaseInv
	fi, err := filter_inventory.New(qChan, cfg, "inventory")
	if err != nil {
		log.Printf("[EE] Failed to create (%s) %s: %v", fi.Typ, fi.Name, err)
		return
	}
	go fi.Run()
	// Dummy file
	cf, err := collector_file.New(qChan, cfg, "file")
	if err != nil {
		log.Printf("[EE] Failed to create (%s) %s: %v", cf.Typ, cf.Name, err)
		return
	}
	go cf.Run()
	// Docker Events
	de, err := qframe_collector_docker_events.New(qChan, cfg, "docker-events")
	if err != nil {
		log.Printf("[EE] Failed to create (%s) %s: %v", de.Typ, de.Name, err)
		return
	}
	go de.Run()
	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}