package prometheus

import (
	"testing"

	"github.com/roidelapluie/o11y-deploy/modules"
)

func TestInterface(*testing.T) {
	var _ modules.ReverseProxiedModule = &Module{}
}
