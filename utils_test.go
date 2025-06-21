package chassis

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPointerOf(t *testing.T) {
	v := 5

	if diff := cmp.Diff(&v, PointerOf(v)); diff != "" {
		t.Errorf("Mismatch in result (-want +got):\n%s", diff)
	}
}
