package core

import (
	"testing"

	"github.com/keadwen/msc_project/proto"
	"github.com/keadwen/msc_project/shared/errdiff"
)

func TestTransmit(t *testing.T) {
	tests := []struct {
		name                string
		src                 *Node
		dst                 *Node
		msg                 int64
		wantSrcEnergy       float64 // Energy after sending data.
		wantDstEnergy       float64 // Energy after receiving data.
		wantSrcDataSent     int64   // Number of packets sent.
		wantDstDataReceived int64   // Nmber of packets received.
		wantDstReceiveQueue int64   // Number of packets in queue.
		wantErr             bool
		wantErrSubstring    string
	}{{
		name: "Distance lower than SQRT(E_FS/E_MP)",
		src: &Node{
			Conf: config.Node{
				Location: &config.Location{}, // (0, 0)
			},
			Ready:  true,
			Energy: 1,
		},
		dst: &Node{
			Conf: config.Node{
				Id: int64(1),
				Location: &config.Location{
					X: float64(10),
					Y: float64(10),
				},
			},
			Ready:  true,
			Energy: 1,
		},
		msg:                 1000,
		wantSrcEnergy:       0.999944,
		wantDstEnergy:       0.999996,
		wantSrcDataSent:     1000,
		wantDstDataReceived: 1000,
		wantDstReceiveQueue: 1000,
	}, {
		name: "Distance greater than SQRT(E_FS/E_MP)",
		src: &Node{
			Conf: config.Node{
				Location: &config.Location{}, // (0, 0)
			},
			Ready:  true,
			Energy: 1,
		},
		dst: &Node{
			Conf: config.Node{
				Id: int64(1),
				Location: &config.Location{
					X: float64(100),
					Y: float64(100),
				},
			},
			Ready:  true,
			Energy: 1,
		},
		msg:                 1000,
		wantSrcEnergy:       0.9958,
		wantDstEnergy:       0.999996,
		wantSrcDataSent:     1000,
		wantDstDataReceived: 1000,
		wantDstReceiveQueue: 1000,
	}, {
		// TODO(keadwen): Re-enable the test once Transmit() returns error.
		name: "Destination node has not enough energy",
		src: &Node{
			Conf: config.Node{
				Location: &config.Location{}, // (0, 0)
			},
			Ready:  true,
			Energy: 1,
		},
		dst: &Node{
			Conf: config.Node{
				Id: int64(1),
				Location: &config.Location{
					X: float64(10),
					Y: float64(10),
				},
			},
			Ready:  true,
			Energy: 0,
		},
		msg:              1000,
		wantSrcEnergy:    0.999944,
		wantSrcDataSent:  1000,
		wantErr:          true,
		wantErrSubstring: "", // "node <1> no more energy!"
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.src.Transmit(tt.msg, tt.dst)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("Transmit() returned diff: %v", diff)
			}
			if tt.src.Energy != tt.wantSrcEnergy {
				t.Fatalf("Transmit() got %v, want %v", tt.src.Energy, tt.wantSrcEnergy)
			}
			if tt.dst.Energy != tt.wantDstEnergy {
				t.Fatalf("Transmit() got %v, want %v", tt.dst.Energy, tt.wantDstEnergy)
			}
			if tt.src.dataSent != tt.wantSrcDataSent {
				t.Fatalf("Transmit() got %v, want %v", tt.src.dataSent, tt.wantSrcDataSent)
			}
			if tt.dst.dataReceived != tt.wantDstDataReceived {
				t.Fatalf("Transmit() got %v, want %v", tt.dst.dataReceived, tt.wantDstDataReceived)
			}
			if tt.dst.receiveQueue != tt.wantDstReceiveQueue {
				t.Fatalf("Transmit() got %v, want %v", tt.dst.dataReceived, tt.wantDstReceiveQueue)
			}
		})
	}
}

func TestReceive(t *testing.T) {
	tests := []struct {
		name             string
		src              *Node
		msg              int64
		wantEnergy       float64
		wantReceiveQueue int64
		wantDataReceived int64
		wantErrSubstring string
	}{{
		name: "Enough energy to receive",
		src: &Node{
			Conf:         config.Node{},
			Ready:        true,
			Energy:       1,
			receiveQueue: 1000,
			dataReceived: 1234,
		},
		msg:              1000,
		wantEnergy:       0.999996,
		wantReceiveQueue: 2000,
		wantDataReceived: 2234,
	}, {
		name: "Not enough energy to receive",
		src: &Node{
			Conf:         config.Node{},
			Ready:        true,
			Energy:       0.000003,
			receiveQueue: 1000,
			dataReceived: 1234,
		},
		msg:              1000,
		wantEnergy:       0,
		wantReceiveQueue: 1000,
		wantDataReceived: 1234,
		wantErrSubstring: "node <0> no more energy!",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.src.Receive(tt.msg, nil)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("Receive() returned diff: %v", diff)
			}
			if tt.src.Energy != tt.wantEnergy {
				t.Fatalf("Receive() got %v, want %v", tt.src.Energy, tt.wantEnergy)
			}
			if tt.src.dataReceived != tt.wantDataReceived {
				t.Fatalf("Receive() got %v, want %v", tt.src.dataReceived, tt.wantDataReceived)
			}
			if tt.src.receiveQueue != tt.wantReceiveQueue {
				t.Fatalf("Receive() got %v, want %v", tt.src.dataReceived, tt.wantReceiveQueue)
			}
		})
	}
}

func TestDistance(t *testing.T) {
	tests := []struct {
		name string
		src  *Node
		dst  *Node
		want float64
	}{{
		name: "Locations are positive values",
		src: &Node{
			Conf: config.Node{
				Location: &config.Location{
					X: float64(0),
					Y: float64(0),
				},
			},
		},
		dst: &Node{
			Conf: config.Node{
				Location: &config.Location{
					X: float64(10),
					Y: float64(10),
				},
			},
		},
		want: 14.142135623730951,
	}, {
		name: "Locations are positive values",
		src: &Node{
			Conf: config.Node{
				Location: &config.Location{
					X: float64(0),
					Y: float64(0),
				},
			},
		},
		dst: &Node{
			Conf: config.Node{
				Location: &config.Location{
					X: float64(-10),
					Y: float64(-10),
				},
			},
		},
		want: 14.142135623730951,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.src.distance(tt.dst); got != tt.want {
				t.Fatalf("distance() got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConsume(t *testing.T) {
	tests := []struct {
		name             string
		src              *Node
		e                float64
		wantReady        bool
		wantEnergy       float64
		wantErrSubstring string
	}{{
		name: "Deducted energy from alive node",
		src: &Node{
			Conf:   config.Node{},
			Ready:  true,
			Energy: 1,
		},
		e:          0.3,
		wantReady:  true,
		wantEnergy: 0.7,
	}, {
		name: "Not enough energy to deduct",
		src: &Node{
			Conf: config.Node{
				Id: int64(1),
			},
			Ready:  true,
			Energy: 1,
		},
		e:                100,
		wantReady:        false,
		wantEnergy:       0,
		wantErrSubstring: "node <1> no more energy!",
	}, {
		name:             "Negative deduction not allowed",
		src:              &Node{},
		e:                -0.5,
		wantErrSubstring: "consume energy cannot be negative: -0.5",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.src.consume(tt.e)
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("consume() returned diff: %v", diff)
			}
			if tt.src.Ready != tt.wantReady {
				t.Fatalf("consume() got %v, want %v", tt.src.Ready, tt.wantReady)
			}
			if tt.src.Energy != tt.wantEnergy {
				t.Fatalf("consume() got %v, want %v", tt.src.Energy, tt.wantEnergy)
			}
		})
	}
}
