package testdown

// Test is an interface for a test.
type Test interface {
	AcceptVisitor(TestVisitor)
}

// TestVisitor performs an operation based on the concrete type of a [Test].
type TestVisitor interface {
	VisitSuite(Suite)
	VisitDocument(Document)
	VisitAssertion(Assertion)
}

// Suite is a [Test] that contains other tests loaded from the same directory.
type Suite struct {
	Name  string
	Dir   string
	Skip  bool
	Tests []Test
}

// AcceptVisitor calls the method on v that corresponds to t's type.
func (t Suite) AcceptVisitor(v TestVisitor) { v.VisitSuite(t) }

// Document describes a Markdown document that (possibly) contains assertions.
type Document struct {
	Name       string
	File       string
	Skip       bool
	Assertions []Assertion
	Errors     []error
}

// AcceptVisitor calls the method on v that corresponds to t's type.
func (t Document) AcceptVisitor(v TestVisitor) { v.VisitDocument(t) }

// Assertion is a [Test] that describes a single assertion within a [Document].
type Assertion struct {
	Name           string
	File           string
	Line           int
	Skip           bool
	InputLanguage  string
	Input          string
	OutputLanguage string
	ExpectedOutput string
}

// AcceptVisitor calls the method on v that corresponds to t's type.
func (t Assertion) AcceptVisitor(v TestVisitor) { v.VisitAssertion(t) }
