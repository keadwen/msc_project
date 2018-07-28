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
	for _, file := range strings.Split(*configFile, ",") {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Failed to open a file %q: %v", file, err)
		}
		conf := &config.Config{}
		if err := proto.UnmarshalText(string(data), conf); err != nil {
			log.Fatalf("Failed to unmarshal config proto %q: %v", file, err)
		}
		if len(conf.GetNodes()) < 1 {
			log.Fatalf("Found 0 nodes in config proto %q", file)
		}

		if err := s.AddScenario(conf.Protocol.String(), conf); err != nil {
			log.Fatalf("Failed to add scenario: %v", err)
		}
		fmt.Printf("File: %s = %v\n", file, conf)
	}

	// Simulation section.
	if err := s.Run(); err != nil {
		log.Fatalf("Failed to run simulation: %v", err)
	}
	if err := s.ExportPlots(fmt.Sprintf("graphs/rounds-%d", time.Now().Nanosecond())); err != nil {
		log.Fatalf("Failed to export plot: %v", err)
	}
}

// createScenario provides an ability to run single scenario with CLI support.
func createScenario() (*config.Config, error) {
	reader := bufio.NewReader(os.Stdin)
	// Get amount of nodes in scenario.
	nodes, err := readIntInput(reader, "Number of nodes in scenario [e.g. 10]: ")
	if err != nil {
		return nil, err
	}
	protocol, err := readIntInput(reader, "Select protcol [0 - ALL, 1 - DIRECT, 2 - LEACH]: ")
	if err != nil {
		return nil, err
	}
	clusters, err := readIntInput(reader, "Number of clusters in scenario [e.g. 3]: ")
	if err != nil {
		return nil, err
	}

	conf := &config.Config{
		Protocol: config.E_Protocol(protocol),
		Clusters: int64(clusters),
	}
	// Create base station (without user interraction).
	conf.Nodes = append(conf.Nodes, &config.Node{
		Id:            0,
		InitialEnergy: 1e100,
		Location: &config.Location{
			X: 500,
			Y: 500,
		},
	})
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

func readIntInput(reader *bufio.Reader, prompt string) (int, error) {
	fmt.Print(prompt)
	input, err := reader.ReadString('\n')
	if err != nil {
		return 0, fmt.Errorf("failed to read input: %v", err)
	}
	value, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil {
		return 0, fmt.Errorf("failed to parse int %q: %v", input, err)
	}
	return value, nil
}
