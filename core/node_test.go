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
		name: "Destination node is not ready",
		src: &Node{
			Conf: config.Node{
				Location: &config.Location{}, // (0, 0)
			},
			Ready:  true,
			Energy: 1,
		},
		dst: &Node{
			Conf: config.Node{
				Location: &config.Location{
					X: float64(10),
					Y: float64(10),
				},
			},
			Ready:  false,
			Energy: 0,
		},
		msg:              1000,
		wantSrcEnergy:    0.999944,
		wantSrcDataSent:  1000,
		wantErr:          true,
		wantErrSubstring: "",
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
				t.Fatalf("Transmit() got %v, want %v", tt.dst.dataReceived, tt.wantDstDataReceived)
			}
		})
	}
}
