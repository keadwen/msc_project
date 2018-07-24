package simulator

import (
	"fmt"

	"github.com/keadwen/msc_project/core"
	"github.com/keadwen/msc_project/proto"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
)

var mapProtocol = map[config.E_Protocol]core.Protocol{
	config.E_Protocol_UNSET:  &core.DirectCommunication{},
	config.E_Protocol_DIRECT: &core.DirectCommunication{},
	config.E_Protocol_LEACH:  &core.LEACH{},
}

type Simulator struct {
	config  *config.Config
	network *core.Network
}

// Create returns a simulation object capable to run a single testing scenario.
func Create(conf *config.Config) (*Simulator, error) {
	s := &Simulator{
		config: conf,
	}
	if err := s.create(); err != nil {
		return nil, err
	}
	return s, nil
}

// Run executes the simulation scenario.
func (s *Simulator) Run() error {
	if s.config == nil {
		return fmt.Errorf("configuration is nil. Did you run Create()?")
	}
	return s.network.Simulate()
}

// ExportPlots create the plot image file.
func (s *Simulator) ExportPlots(filepath string) error {
	if filepath == "" {
		return fmt.Errorf("empty filepath provided")
	}
	var err error
	err = s.network.PlotRound.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("%s-%s-round.png", filepath, s.config.Protocol.String()))
	err = s.network.PlotNodes.Save(8*vg.Inch, 8*vg.Inch, fmt.Sprintf("%s-%s-nodes.png", filepath, s.config.Protocol.String()))
	return err
}

// create builds the simulator object according to a given configuration.
func (s *Simulator) create() error {
	// Create a plot object, recording each simulation round.
	pRound, err := createPlot("Total energy in rounds", "Round", "Total energy [J]")
	if err != nil {
		return fmt.Errorf("failed to create plot object: %v", err)
	}
	pNodes, err := createPlot("Number of nodes in rounds", "Round", "Number of nodes")
	if err != nil {
		return fmt.Errorf("failed to create plot object: %v", err)
	}

	// Select the protocol in simulation.
	protocol := mapProtocol[s.config.GetProtocol()]
	if protocol == nil {
		return fmt.Errorf("failed to match a protocol: %v", s.config.GetProtocol())
	}
	protocol.SetClusters(int(s.config.Clusters))
	protocol.SetNodes(len(s.config.GetNodes()) - 1) // Do not count base station.

	// Create new network space.
	s.network = &core.Network{
		Protocol:  protocol,
		PlotRound: pRound,
		PlotNodes: pNodes,
	}

	// Create base station node, which is node with ID of 0.
	n := s.config.Nodes[0]
	s.network.BaseStation = &core.Node{
		Conf:   *n,
		Ready:  true,
		Energy: n.GetInitialEnergy(),
	}

	// Create the rest of the nodes.
	for i := 1; i < len(s.config.GetNodes()); i++ {
		n := s.config.Nodes[i]
		s.network.AddNode(&core.Node{
			Conf:   *n,
			Ready:  true,
			Energy: n.GetInitialEnergy(),
		})
	}
	return nil
}

// createPlot returns new plot.Plot object.
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
