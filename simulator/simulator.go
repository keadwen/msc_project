package simulator

import (
	"fmt"
	"sort"

	"github.com/keadwen/msc_project/core"
	"github.com/keadwen/msc_project/proto"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

var mapProtocol = map[config.E_Protocol]core.Protocol{
	config.E_Protocol_UNSET:  &core.DirectCommunication{},
	config.E_Protocol_DIRECT: &core.DirectCommunication{},
	config.E_Protocol_LEACH:  &core.LEACH{},
	// config.E_Protocol_APTEEN: &core.APTEEN{},
	config.E_Protocol_PEGASIS: &core.PEGASIS{},
}

type Simulator struct {
	namespace map[string]bool
	config    map[string]*config.Config
	network   map[string]*core.Network

	plotTotalEnergy *plot.Plot // An amount of total energy in the network per round.
	plotNodes       *plot.Plot // A number of alive nodes in the network per round.
}

// Create returns a simulation object capable to run a multiple testing scenario.
func Create() (*Simulator, error) {
	// Create a plot objects, recording each simulation on single graph.
	pte, err := createPlot("Total energy in the network", "Round", "Total energy [J]")
	if err != nil {
		return nil, fmt.Errorf("failed to create plot object: %v", err)
	}
	pn, err := createPlot("Number of alive nodes in the network", "Round", "Number of nodes")
	if err != nil {
		return nil, fmt.Errorf("failed to create plot object: %v", err)
	}
	s := &Simulator{
		namespace:       map[string]bool{},
		config:          map[string]*config.Config{},
		network:         map[string]*core.Network{},
		plotTotalEnergy: pte,
		plotNodes:       pn,
	}
	return s, nil
}

// AddScenario adds a new testing scenario.
func (s *Simulator) AddScenario(name string, conf *config.Config) error {
	if _, exists := s.namespace[name]; exists {
		return fmt.Errorf("scenario %s already exists", name)
	}
	s.namespace[name] = true
	s.config[name] = conf
	return s.create(name, conf)
}

// Run executes all simulation scenario.
func (s *Simulator) Run() error {
	if len(s.config) == 0 {
		return fmt.Errorf("empty config map. Did you add a testing scenario?")
	}

	var names []string
	for name := range s.namespace {
		names = append(names, name)
	}
	sort.Strings(names)

	var err error
	for _, name := range names {
		fmt.Printf("== Running simulation %q ==\n", name)
		err = s.network[name].Simulate()
		if err != nil {
			return err
		}
	}
	return nil
}

// ExportPlots create the plot image file.
func (s *Simulator) ExportPlots(filepath string) error {
	if filepath == "" {
		return fmt.Errorf("empty filepath provided")
	}
	var err error
	err = s.plotter()
	err = s.plotTotalEnergy.Save(16*vg.Inch, 16*vg.Inch, fmt.Sprintf("%s-round.png", filepath))
	err = s.plotNodes.Save(16*vg.Inch, 16*vg.Inch, fmt.Sprintf("%s-nodes.png", filepath))
	return err
}

// create builds the simulator object according to a given configuration.
func (s *Simulator) create(name string, conf *config.Config) error {
	// Select the protocol in simulation.
	protocol := mapProtocol[conf.GetProtocol()]
	if protocol == nil {
		return fmt.Errorf("failed to match a protocol: %v", conf.GetProtocol())
	}
	nodes := len(conf.GetNodes()) - 1 // Do not count base station.
	protocol.SetClusters(int(float64(nodes) * conf.PClusterHeads))
	protocol.SetNodes(nodes)

	// Create new network space.
	s.network[name] = &core.Network{
		Protocol:        protocol,
		MaxRounds:       conf.MaxRounds,
		PlotTotalEnergy: s.plotTotalEnergy,
		PlotNodes:       s.plotNodes,
	}

	// Create base station node, which is node with ID of 0.
	n := conf.Nodes[0]
	s.network[name].BaseStation = &core.Node{
		Conf:   *n,
		Ready:  true,
		Energy: n.GetInitialEnergy(),
	}

	// Create the rest of the nodes.
	for i := 1; i < len(conf.GetNodes()); i++ {
		n := conf.Nodes[i]
		s.network[name].AddNode(&core.Node{
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

// TODO(keadwen): Rework the function to scale with additonal algorithms.
func (s *Simulator) plotter() error {
	var pna []interface{}
	for name, net := range s.network {
		pna = append(pna, name, net.NodesAlivePoints)
	}
	if err := plotutil.AddLines(s.plotNodes, pna...); err != nil {
		return fmt.Errorf("failed to AddLines(): %v", err)
	}

	var pne []interface{}
	for name, net := range s.network {
		networkEnergy := plotter.XYs{}
		for r := 1; r < int(net.Round); r++ {
			var e float64
			net.Nodes.Range(func(_, n interface{}) bool {
				e += net.NodesEnergyPoints[n.(*core.Node).Conf.GetId()][r].Y
				return true
			})
			networkEnergy = append(networkEnergy, plotter.XYs{{X: float64(r), Y: e}}...)
		}
		pne = append(pne, name, networkEnergy)
	}
	if err := plotutil.AddLines(s.plotTotalEnergy, pne...); err != nil {
		return fmt.Errorf("failed to AddLines(): %v", err)
	}
	return nil
}
