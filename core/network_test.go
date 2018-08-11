package core

import (
	"testing"

	"github.com/keadwen/msc_project/shared/errdiff"
)

func TestSimulate(t *testing.T) {
	tests := []struct {
		name             string
		p                Protocol
		net              *Network
		maxRounds        int64
		wantRX           int64
		wantTX           int64
		wantErrSubstring string
	}{{
		name:      "Valid LEACH",
		p:         &LEACH{3, 11},
		net:       makeNetwork(),
		maxRounds: 10,
		wantRX:    11000,
		wantTX:    1000,
	}, {
		name:      "Valid Direct Communication",
		p:         &DirectCommunication{0, 11},
		net:       makeNetwork(),
		maxRounds: 10,
		wantRX:    11000,
		wantTX:    1000,
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.net.Protocol = tt.p
			tt.net.MaxRounds = tt.maxRounds

			err := tt.net.Simulate()
			if diff := errdiff.Substring(err, tt.wantErrSubstring); diff != "" {
				t.Fatalf("Simulate() returned diff: %v", diff)
			}
			if got := tt.net.BaseStation.dataReceived; got != tt.wantRX {
				t.Fatalf("Simulate() got %v, want %v", got, tt.wantRX)
			}
			tt.net.Nodes.Range(func(_, n interface{}) bool {
				if got := n.(*Node).dataSent; got < tt.wantTX {
					t.Fatalf("Simulate() node<%+v> got %v, want >= %v", n.(*Node), got, tt.wantTX)
				}
				return true
			})
		})
	}
}
