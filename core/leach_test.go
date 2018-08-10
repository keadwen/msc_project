package core

import (
	"testing"

	"github.com/keadwen/msc_project/proto"
	"github.com/keadwen/msc_project/shared/errdiff"
)

func makeNetwork() *Network {
	net := &Network{}

	// Create configuration.
	nodes := []*config.Node{
		{Id: 0, InitialEnergy: 1e100, Location: &config.Location{X: float64(200), Y: float64(0)}}, // Base station.
		// Column X: 100
		{Id: 1, InitialEnergy: 1.0, Location: &config.Location{X: float64(100), Y: float64(100)}}, //
		{Id: 2, InitialEnergy: 1.0, Location: &config.Location{X: float64(100), Y: float64(200)}},
		{Id: 3, InitialEnergy: 1.0, Location: &config.Location{X: float64(100), Y: float64(300)}},
		// Column X: 150
		{Id: 4, InitialEnergy: 1.0, Location: &config.Location{X: float64(150), Y: float64(200)}},
		// Column X: 200
		{Id: 5, InitialEnergy: 1.0, Location: &config.Location{X: float64(200), Y: float64(100)}},
		{Id: 6, InitialEnergy: 1.0, Location: &config.Location{X: float64(200), Y: float64(200)}},
		{Id: 7, InitialEnergy: 1.0, Location: &config.Location{X: float64(200), Y: float64(300)}},
		// Column X: 250
		{Id: 8, InitialEnergy: 1.0, Location: &config.Location{X: float64(250), Y: float64(200)}},
		// Column X: 300
		{Id: 9, InitialEnergy: 1.0, Location: &config.Location{X: float64(300), Y: float64(100)}},
		{Id: 10, InitialEnergy: 1.0, Location: &config.Location{X: float64(300), Y: float64(200)}},
		{Id: 11, InitialEnergy: 1.0, Location: &config.Location{X: float64(300), Y: float64(300)}},
	}

	// Create base station node, which is node with ID of 0.
	n := nodes[0]
	net.BaseStation = &Node{
		Conf:   *n,
		Ready:  true,
		Energy: n.GetInitialEnergy(),
	}

	// Create the rest of the nodes.
	for i := 1; i < len(nodes); i++ {
		n := nodes[i]
		net.AddNode(&Node{
			Conf:   *n,
			Ready:  true,
			Energy: n.GetInitialEnergy(),
		})
	}

	return net
}

func TestSetup(t *testing.T) {
	tests := []struct {
		name             string
		p                Protocol
		net              *Network
		wantClusters     int
		wantErrSubstring string
	}{{
		name: "Valid Direct Communication",
		p: &DirectCommunication{
			Nodes:    10,
			Clusters: 0,
		},
		net:          makeNetwork(),
		wantClusters: 0,
	}, {
		name: "Valid LEACH",
		p: &LEACH{
			Nodes:    10,
			Clusters: 2,
		},
		net:          makeNetwork(),
		wantClusters: 2,
	}}

	for _, tt := range tests {
		heads, err := tt.p.Setup(tt.net)
		if len(heads) != tt.wantClusters {
			t.Fatalf("Setup() got %v heads, want %v", len(heads), tt.wantClusters)
		}
		if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
			t.Fatalf("Setup() returned diff: %v", diff)
		}

		// Test if each node has a valid next hop.
		validHeads := append(heads, 0) // Add base station as a valid next hop.
		tt.net.Nodes.Range(func(_, n interface{}) bool {
			for _, id := range validHeads {
				if n.(*Node).nextHop.Conf.GetId() == id {
					return true // nodes next hop is valid, move to next node.
				}
			}
			// Nodes next hop not in a list of valid cluster heads.
			t.Fatalf("Setup() got %v nextHop ID, want one of %v", n.(*Node).nextHop.Conf.GetId(), validHeads)
			return true
		})
	}
}
