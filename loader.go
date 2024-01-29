package testdown

import (
	"bytes"
	"context"
	"fmt"
	"io/fs"
	"path"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

// Loader loads [Test] values from (directories containing) Markdown documents.
type Loader struct {
	FS     fs.FS
	Parser parser.Parser
}

// Load returns tests loaded from the file or directory at the given path.
func (l *Loader) Load(ctx context.Context, p string) (Test, error) {
	if ctx.Err() != nil {
		return Suite{}, ctx.Err()
	}

	info, err := fs.Stat(l.FS, p)
	if err != nil {
		return nil, err
	}

	dir := path.Dir(p)

	if info.IsDir() {
		return l.loadSuite(ctx, dir, info.Name())
	}

	return l.loadDocument(ctx, dir, info.Name())
}

func (l *Loader) loadSuite(ctx context.Context, dir, name string) (Suite, error) {
	if ctx.Err() != nil {
		return Suite{}, ctx.Err()
	}

	qual := path.Join(dir, name)
	name, skip := strings.CutPrefix(name, "_")

	entries, err := fs.ReadDir(l.FS, qual)
	if err != nil {
		return Suite{}, err
	}

	test := Suite{
		Name: name,
		Dir:  qual,
		Skip: skip,
	}

	for _, entry := range entries {
		var sub Test

		if entry.IsDir() {
			sub, err = l.loadSuite(ctx, qual, entry.Name())
		} else if l.isTestDocument(entry.Name()) {
			sub, err = l.loadDocument(ctx, qual, entry.Name())
		} else {
			continue
		}

		if err != nil {
			return Suite{}, err
		}

		test.Tests = append(test.Tests, sub)
	}

	return test, nil
}

func (l *Loader) loadDocument(ctx context.Context, dir, name string) (Document, error) {
	if ctx.Err() != nil {
		return Document{}, ctx.Err()
	}

	qual := path.Join(dir, name)
	name, skip := strings.CutPrefix(name, "_")

	if !l.isTestDocument(qual) {
		return Document{}, fmt.Errorf("%s is not a testdown document", qual)
	}

	source, err := fs.ReadFile(l.FS, qual)
	if err != nil {
		return Document{}, err
	}

	parser := l.Parser
	if parser == nil {
		parser = goldmark.DefaultParser()
	}
	root := parser.Parse(text.NewReader(source))

	test := Document{
		Name: name,
		File: qual,
		Skip: skip,
	}

	var assertion *Assertion

	return test, ast.Walk(
		root,
		func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			if !entering {
				return ast.WalkContinue, nil
			}

			code, ok := n.(*ast.FencedCodeBlock)
			if !ok {
				return ast.WalkContinue, nil
			}

			info := parseCodeInfo(code, source)
			line := lineNumberOf(code, source)

			if !info.IsAssertion {
				assertion = &Assertion{
					Name:          "L" + strconv.Itoa(line),
					File:          qual,
					Line:          line,
					InputLanguage: info.Language,
					Input:         sourceOf(code, source),
				}
				return ast.WalkContinue, nil
			} else if assertion == nil {
				test.Errors = append(
					test.Errors,
					fmt.Errorf(
						"found testdown assertion at %s:%d without a preceding code block",
						qual,
						line,
					),
				)
			} else {
				assertion.OutputLanguage = info.Language
				assertion.ExpectedOutput = sourceOf(code, source)
				assertion.Skip = info.Skip
				test.Assertions = append(test.Assertions, *assertion)
				assertion = nil
			}

			return ast.WalkContinue, nil
		},
	)
}

// isTestDocument returns true if the given file should be treated as a Markdown
// document that (potentially) contains test assertions.
func (l *Loader) isTestDocument(filename string) bool {
	return strings.HasSuffix(filename, ".testdown.md")
}

// sourceOf returns the Markdown source for n.
func sourceOf(n ast.Node, source []byte) string {
	text := string(n.Text(source))
	lines := n.Lines()

	for i := 0; i < lines.Len(); i++ {
		line := lines.At(i)
		text += string(line.Value(source))
	}

	return text
}

// codeInfo is the result of parsing the "info line" of a code block.
type codeInfo struct {
	Language    string
	IsAssertion bool
	Skip        bool
}

// / parseCodeInfo parses the "info line" (the line indicating the block's
// language) of a code block.
func parseCodeInfo(n *ast.FencedCodeBlock, source []byte) codeInfo {
	info := codeInfo{
		Language: "text",
	}

	if n.Info == nil {
		return info
	}

	text := string(n.Info.Segment.Value(source))
	fields := strings.Fields(text)

	for i, field := range fields {
		if field == "testdown" {
			info.IsAssertion = true
		} else if i == 0 {
			info.Language = field
		} else if field == "skip" && info.IsAssertion {
			info.Skip = true
		}
	}

	return info
}

// lineNumberOf returns the first line number of n.
func lineNumberOf(n ast.Node, source []byte) int {
	i := n.Lines().At(0).Start
	return 1 + bytes.Count(source[:i], []byte("\n"))
}
