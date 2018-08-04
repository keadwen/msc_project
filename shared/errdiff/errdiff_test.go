package errdiff

import (
	"fmt"
	"testing"
)

func TestSubstring(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		substring string
		want      string
	}{{
		name: "Empty diff (no error, no substring)",
		want: "",
	}, {
		name:      "Empty diff (error matches substring)",
		err:       fmt.Errorf("foo"),
		substring: "fo",
		want:      "",
	}, {
		name: "Diff (error given, no substring)",
		err:  fmt.Errorf("foo"),
		want: "got err=foo, want err=nil",
	}, {
		name:      "Diff (error given, substring does not match",
		err:       fmt.Errorf("foo"),
		substring: "bar",
		want:      "got err=foo, want err containing bar",
	}, {
		name:      "Diff (no error, substring does not match",
		substring: "bar",
		want:      "got err=nil, want err containing bar",
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Substring(tt.err, tt.substring); got != tt.want {
				t.Fatalf("Substring() got %v, want %v", got, tt.want)
			}
		})
	}
}
