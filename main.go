package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/keadwen/msc_project/proto"
	"github.com/keadwen/msc_project/simulator"
)

var (
	configFile = flag.String("config_files", "", "location of node config files (comma separated)")
)

func main() {
	flag.Parse()

	// Create Simulator.
	s, err := simulator.Create()
	if err != nil {
		log.Fatalf("Failed to create simulator: %v", err)
	}

	// Parse or create simulation configuration proto.
	configs := []*config.Config{}
	if *configFile == "" {
		configs = defaultConfigs()
	} else {
		for _, file := range strings.Split(*configFile, ",") {
			data, err := ioutil.ReadFile(file)
			if err != nil {
				log.Fatalf("Failed to open a file %q: %v", file, err)
			}
			conf := &config.Config{}
			if err := proto.UnmarshalText(string(data), conf); err != nil {
				log.Fatalf("Failed to unmarshal config proto %q: %v", file, err)
			}
			if len(conf.GetNodes()) < 2 {
				log.Fatalf("Found %d nodes in config proto %q. Expectes more than 1.", len(conf.GetNodes()), file)
			}
			configs = append(configs, conf)
		}
	}

	// Simulation section.
	for id, conf := range configs {
		name := fmt.Sprintf("%s-%v-%v-%d", conf.Protocol.String(), len(conf.GetNodes())-1, conf.Nodes[1].GetInitialEnergy(), id)
		if err := s.AddScenario(name, conf); err != nil {
			log.Fatalf("Failed to add scenario: %v", err)
		}
		fmt.Printf("Added scenario: %s\n", name)
	}
	if err := s.Run(); err != nil {
		log.Fatalf("Failed to run simulation: %v", err)
	}
	if err := s.ExportPlots(fmt.Sprintf("graphs/%d", time.Now().Nanosecond())); err != nil {
		log.Fatalf("Failed to export plot: %v", err)
	}
}

func defaultConfigs() []*config.Config {
	var configs []*config.Config
	rand.Seed(time.Now().UTC().UnixNano())

	nodes := 150
	sizes := []float64{1000} // , 200, 250, 300, 350, 400}
	energies := []float64{10.0}
	protocols := []int{1, 2}
	for _, size := range sizes {
		// Nodes in a single network size, must have same location.
		loc := []float64{size / 2} // Default base station location.
		for n := 1; n <= nodes; n++ {
			loc = append(loc, rand.Float64()*size)
		}

		var tmp []*config.Config
		for _ = range energies {
			for _, protocol := range protocols {
				conf := &config.Config{
					Protocol: config.E_Protocol(protocol),
					Clusters: int64(nodes / 10),
				}
				// Create base station.
				conf.Nodes = append(conf.Nodes, &config.Node{
					Id:            0,
					InitialEnergy: 1e100,
					Location: &config.Location{
						X: loc[0],
						Y: loc[0],
					},
				})
				tmp = append(tmp, conf)
			}
		}

		for _, conf := range tmp {
			for _, energy := range energies {
				for n := 1; n <= nodes; n++ {
					conf.Nodes = append(conf.Nodes, &config.Node{
						Id:            int64(n),
						InitialEnergy: energy,
						Location: &config.Location{
							X: loc[n],
							Y: loc[n],
						},
					})

				}
			}
		}
		configs = append(configs, tmp...)
	}

	return configs
}
