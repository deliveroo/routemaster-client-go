package suite

import "github.com/deliveroo/routemaster-client-go/integrationtest"

func ConfigureRunner(runner *integrationtest.Runner) {
	for _, t := range allTests {
		runner.AppendTest(t)
	}
}
