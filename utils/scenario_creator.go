package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/keadwen/msc_project/proto"
)

var (
	nodesNumber      = flag.Int("nodes_number", 200, "Number of nodes in scenario")
	clusterPercent   = flag.Float64("cluster_percent", 0.1, "Percent value of clusters per round <0, 1>")
	initialEnergy    = flag.Float64("initial_energy", 0.5, "Initial nodes energy in [J]")
	protocol         = flag.Int64("protocol", 1, "As default uses Direct Communication")
	areaEdge         = flag.Float64("area_edge", 200.0, "Size of an edge in square region")
	baseStationXAxis = flag.Float64("base_station_x_axis", 100.0, "Coordinate X of base station")
	baseStationYAxis = flag.Float64("base_station_y_axis", 100.0, "Coordinate Y of base station")
	outputFile       = flag.String("outputFile", "", "As default uses stdout")
)

func main() {
	flag.Parse()

	cfg := &config.Config{
		Protocol: config.E_Protocol(*protocol),
		Clusters: int64(*clusterPercent * float64(*nodesNumber)),
	}
	// Create base station
	cfg.Nodes = append(cfg.Nodes, &config.Node{
		Id:            0,
		InitialEnergy: 1e100,
		Location: &config.Location{
			X: *baseStationXAxis,
			Y: *baseStationYAxis,
		},
	})

	// Create nodes.
	rand.Seed(time.Now().UTC().UnixNano())
	for id := 1; id <= *nodesNumber; id++ {
		cfg.Nodes = append(cfg.Nodes, &config.Node{
			Id:            int64(id),
			InitialEnergy: *initialEnergy,
			Location: &config.Location{
				X: *areaEdge * rand.Float64(),
				Y: *areaEdge * rand.Float64(),
			},
		})
	}

	// Save to file or display in stdout.
	if *outputFile == "" {
		fmt.Println(proto.MarshalTextString(cfg))
	} else {
		f, err := os.Create(*outputFile)
		if err != nil {
			fmt.Println("Failed to create file %q: %v", *outputFile, err)
		}
		w := bufio.NewWriter(f)
		w.WriteString(proto.MarshalTextString(cfg))
		w.Flush()
	}
}
