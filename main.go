package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/keadwen/msc_project/proto"
	"github.com/keadwen/msc_project/simulator"
)

var (
	configFile = flag.String("config_file", "", "location of node config file")
)

func main() {
	flag.Parse()

	// Parse or create simulation configuration proto.
	conf := &config.Config{}
	var err error
	if *configFile == "" {
		conf, err = createScenario()
		if err != nil {
			log.Fatalf("Failed to create scenario: %v", err)
		}
	} else {
		data, err := ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("Failed to open a file %q: %v", *configFile, err)
		}
		if err := proto.UnmarshalText(string(data), conf); err != nil {
			log.Fatalf("Failed to unmarshal config proto %q: %v", *configFile, err)
		}
	}
	if len(conf.GetNodes()) < 1 {
		log.Fatalf("Found 0 nodes in config proto %q", *configFile)
	}

	// Simulation section.
	s, err := simulator.Create(conf)
	if err != nil {
		log.Fatalf("Failed to create simulation: %v", err)
	}
	if err := s.Run(); err != nil {
		log.Fatalf("Failed to run simulation: %v", err)
	}
	if err := s.ExportPlot(fmt.Sprintf("graphs/rounds-%d.png", time.Now().Nanosecond())); err != nil {
		log.Fatalf("Failed to export plot: %v", err)
	}
}

// createScenario provides an ability to run single scenario with CLI support.
func createScenario() (*config.Config, error) {
	reader := bufio.NewReader(os.Stdin)
	// Get amount of nodes in scenario.
	fmt.Print("Nodes in scenario [e.g. 5]: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("failed to read input: %v", err)
	}
	nodes, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return nil, fmt.Errorf("failed to parse int %q: %v", input, err)
	}

	conf := &config.Config{}
	// Create base station (without user interraction).
	conf.Nodes = append(conf.Nodes, &config.Node{Id: 0, InitialEnergy: 1e100, Location: &config.Location{X: 500, Y: 500}})
	fmt.Println("Created Base Station...")

	// Create nodes with exact amount of energy, but random location [0, 1000].
	for n := 1; n <= nodes; n++ {
		conf.Nodes = append(conf.Nodes, &config.Node{
			Id:            int64(n),
			InitialEnergy: 0.1,
			Location: &config.Location{
				X: rand.Float64() * 1000,
				Y: rand.Float64() * 1000,
			},
		})
		node := conf.Nodes[n]
		fmt.Printf("Created node <%d> (E: %f[J] X: %2.f, Y: %2.f)...\n", n, node.GetInitialEnergy(), node.GetLocation().GetX(), node.GetLocation().GetY())
	}
	return conf, nil
}
