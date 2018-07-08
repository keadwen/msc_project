package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
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

	var conf config.Config
	var err error
	if *configFile == "" {
		conf, err = createScenario()
	} else {
		data, err := ioutil.ReadFile(*configFile)
		if err != nil {
			log.Fatalf("Failed to open a file %q: %v", *configFile, err)
		}
		if err := proto.UnmarshalText(string(data), &conf); err != nil {
			log.Fatalf("Failed to unmarshal config proto %q: %v", *configFile, err)
		}
	}
	if len(conf.GetNodes()) < 1 {
		log.Fatalf("Found 0 nodes in config proto %q", *configFile)
	}

	// Create a plot for rounds.
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Plotutil example"
	p.X.Label.Text = "Round"
	p.Y.Label.Text = "Energy"

	// Create new space.
	net := &core.Network{
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
	if err := p.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("graphs/rounds-%d.png", time.Now().Nanosecond())); err != nil {
		panic(err)
	}
}

// createScenario provides an ability to run single scenario with CLI support.
func createScenario() (config.Config, error) {
	return config.Config{}, nil
}
