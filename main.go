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
	"github.com/keadwen/msc_project/core"
	"github.com/keadwen/msc_project/proto"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
)

var (
	configFile = flag.String("config_file", "", "location of node config file")
)

func main() {
	flag.Parse()

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

	// Create a plot for rounds.
	p, err := createPlot("energy(rounds)", "round", "energy [J]")
	if err != nil {
		log.Fatalf("Failed to create plot object: %v", err)
	}

	// Create new space.
	net := &core.Network{
		Protocol: &core.DirectCommunication{},
		//Protocol:  &core.LEACH{1, len(conf.GetNodes()) - 1},
		PlotRound: p,
	}

	// Create base station node.
	bsc := conf.Nodes[0]
	net.BaseStation = &core.Node{
		Conf:   *bsc,
		Ready:  true,
		Energy: bsc.GetInitialEnergy(),
	}

	// Create nodes.
	for n := 1; n < len(conf.GetNodes()); n++ {
		net.AddNode(&core.Node{Conf: *conf.Nodes[n], Ready: true})
	}

	// Run simulation for 10 rounds.
	net.Simulate()

	// Save the plot to a PNG file.
	pif := fmt.Sprintf("graphs/rounds-%d.png", time.Now().Nanosecond())
	if err := p.Save(8*vg.Inch, 8*vg.Inch, pif); err != nil {
		log.Fatalf("Failed to save a plot image file %q: %v", pif, err)
	}
}

// createPlot returns new plot object.
func createPlot(title, x, y string) (*plot.Plot, error) {
	p, err := plot.New()
	if err != nil {
		return nil, err
	}
	p.Title.Text = title
	p.X.Label.Text = x
	p.Y.Label.Text = y
	return p, err
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
	conf.Nodes = append(conf.Nodes, &config.Node{Id: 0, InitialEnergy: 1e100, Location: &config.Location{X: 0, Y: 0}})
	fmt.Println("Created Base Station...")

	// Create nodes with exact amount of energy, but random location [0, 1000].
	for n := 1; n <= nodes; n++ {
		conf.Nodes = append(conf.Nodes, &config.Node{
			Id:            int64(n),
			InitialEnergy: 1e-4,
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
