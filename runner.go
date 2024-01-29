package testdown

import (
	"testing"

	"github.com/andreyvit/diff"
)

// TestingT is a constraint for a type that is compatible with [testing.T].
type TestingT[T any] interface {
	Helper()
	Parallel()
	Run(string, func(T)) bool
	Logf(string, ...any)
	Errorf(string, ...any)
	Fatalf(string, ...any)
	SkipNow()
}

// NativeRunner runs testdown tests within Go's native testing framework.
type NativeRunner = Runner[*testing.T]

// Runner executes testdown tests within any test framework with an interface
// similar to Go's native testing framework.
type Runner[T TestingT[T]] struct {
	Output func(Assertion) (string, error)
}

// Run executes performs the assertions within the given [Test] using t.
func (r *Runner[T]) Run(t T, test Test) {
	t.Helper()
	test.AcceptVisitor(&runner[T]{
		Runner:   r,
		TestingT: t,
	})
}

type runner[T TestingT[T]] struct {
	Runner   *Runner[T]
	TestingT T
}

func (r *runner[T]) VisitSuite(s Suite) {
	r.TestingT.Helper()
	r.TestingT.Run(
		s.Name,
		func(t T) {
			t.Parallel()

			if s.Skip {
				t.SkipNow()
				return
			}

			for _, sub := range s.Tests {
				r.Runner.Run(t, sub)
			}
		},
	)
}

func (r *runner[T]) VisitDocument(d Document) {
	r.TestingT.Helper()
	r.TestingT.Run(
		d.Name,
		func(t T) {
			t.Parallel()

			if d.Skip {
				t.SkipNow()
				return
			}

			for _, a := range d.Assertions {
				r.Runner.Run(t, a)
			}
		},
	)
}

func (r *runner[T]) VisitAssertion(a Assertion) {
	r.TestingT.Helper()
	r.TestingT.Run(
		a.Name,
		func(t T) {
			// NOTE: not parallel, we want to see the results within a single
			// document in order
			if a.Skip {
				t.SkipNow()
				return
			}

			t.Logf(
				"--- INPUT (%s) ---\n%s",
				a.InputLanguage,
				a.Input,
			)

			output, err := r.Runner.Output(a)
			if err != nil {
				t.Fatalf("--- OUTPUT (error) ---\n%s", err)
			} else if output != a.ExpectedOutput {
				t.Fatalf(
					"--- OUTPUT (%s, -want +got) ---\n%s",
					a.OutputLanguage,
					diff.LineDiff(a.ExpectedOutput, output),
				)
			} else {
				t.Logf(
					"--- OUTPUT (%s) ---\n%s",
					a.OutputLanguage,
					output,
				)
			}
		},
	)
}
