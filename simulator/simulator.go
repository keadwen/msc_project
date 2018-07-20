package simulator

import (
	"fmt"

	"github.com/keadwen/msc_project/core"
	"github.com/keadwen/msc_project/proto"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/vg"
)

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
	s.network.Simulate()
	return nil
}

// ExportPlot create the plot image file.
func (s *Simulator) ExportPlot(filepath string) error {
	if filepath == "" {
		return fmt.Errorf("empty filepath provided")
	}
	return s.network.PlotRound.Save(8*vg.Inch, 8*vg.Inch, filepath)
}

// create builds the simulator object according to a given configuration.
func (s *Simulator) create() error {
	// Create a plot object, recording each simulation round.
	p, err := createPlot("Total energy in rounds", "round", "Total energy [J]")
	if err != nil {
		fmt.Errorf("failed to create plot object: %v", err)
	}

	// Create new network space.
	// TODO(keadwen): Find a way to pass protocols.
	s.network = &core.Network{
		Protocol: &core.DirectCommunication{},
		// Protocol:  &core.LEACH{1, len(conf.GetNodes()) - 1},
		PlotRound: p,
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
