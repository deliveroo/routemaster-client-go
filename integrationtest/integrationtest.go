package integrationtest

import (
	"fmt"
	"time"

	routemaster "github.com/deliveroo/routemaster-client-go"
)

// Test represents an individual test.
type Test struct {
	Name string
	Func func(*T)
}

type Runner struct {
	URL            string
	RootToken      string
	beforeEachTest []func()
	tests          []*Test
}

func (r *Runner) BeforeEachTest(f func()) {
	r.beforeEachTest = append(r.beforeEachTest, f)
}

func (r *Runner) AppendTest(t *Test) {
	r.tests = append(r.tests, t)
}

func (r *Runner) RunTests() {
	for _, test := range r.tests {
		t := &T{
			name:      test.Name,
			created:   time.Now(),
			url:       r.URL,
			rootToken: r.RootToken,
		}
		test.Func(t)
		t.printOutcome()
	}
}

// T is a helper object for a test run.
type T struct {
	created   time.Time
	name      string
	errors    []string
	warnings  []string
	skipped   bool
	url       string
	rootToken string
}

// NewClient returns a new client with the specified uuid.
func (t *T) NewClient(uuid string) *routemaster.Client {
	client, err := routemaster.NewClient(&routemaster.Config{
		URL:  t.url,
		UUID: uuid,
	})
	if err != nil {
		panic(err)
	}
	return client
}

// RootClient returns a new client with root access.
func (t *T) RootClient() *routemaster.Client {
	return t.NewClient(t.rootToken)
}

// Warn records a warning in a test.
func (t *T) Warn(msg string) {
	t.warnings = append(t.warnings, msg)
}

// Warnf records a warning in a test.
func (t *T) Warnf(format string, args ...interface{}) {
	t.Warn(fmt.Sprintf(format, args))
}

// Error records an error in a test.
func (t *T) Error(msg string) {
	t.errors = append(t.errors, msg)
}

// Errorf records an error in a test.
func (t *T) Errorf(format string, args ...interface{}) {
	t.Error(fmt.Sprintf(format, args))
}

// Fatal records an error in a test and aborts the rest of the test.
func (t *T) Fatal(msg string) {
	panic(msg)
}

// Fatalf records an error in a test and aborts the rest of the test.
func (t *T) Fatalf(format string, args ...interface{}) {
	t.Fatal(fmt.Sprintf(format, args))
}

func (t *T) printOutcome() {
	elapsed := time.Since(t.created)
	outcome := "ok"
	if t.skipped {
		outcome = "SKIP"
	} else if len(t.errors) != 0 {
		outcome = "FAIL"
	}
	fmt.Printf("%-8s %-60s %s\n", outcome, t.name, elapsed)
	for i, msg := range t.warnings {
		fmt.Printf("(%d) warn: %s\n", i, msg)
	}
	for i, msg := range t.errors {
		fmt.Printf("(%d) %s\n", i, msg)
	}
}
