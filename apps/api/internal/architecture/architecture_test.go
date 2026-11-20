// Package architecture holds the tests that keep the dependency rules of the
// API true: the bounded contexts never import each other, and dependencies
// point inward. It reads the imports of the sources, it runs nothing.
package architecture_test

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// internalRoot is apps/api/internal, seen from this package's directory.
const internalRoot = ".."

// thirdPartyInDomain is what a domain may import besides the standard library.
var thirdPartyInDomain = map[string]bool{"github.com/google/uuid": true}

// source is a non-test Go file and what it imports.
type source struct {
	path    string // slash-separated, relative to internal/
	imports []string
}

func modulePath(t *testing.T) string {
	t.Helper()
	gomod, err := os.ReadFile(filepath.Join(internalRoot, "..", "go.mod"))
	require.NoError(t, err)
	for _, line := range strings.Split(string(gomod), "\n") {
		if module, ok := strings.CutPrefix(line, "module "); ok {
			return strings.TrimSpace(module)
		}
	}
	require.Fail(t, "go.mod has no module line")
	return ""
}

func loadSources(t *testing.T) []source {
	t.Helper()
	var sources []source
	err := filepath.WalkDir(internalRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "architecture" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(internalRoot, path)
		if err != nil {
			return err
		}
		s := source{path: filepath.ToSlash(rel)}
		for _, spec := range file.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			s.imports = append(s.imports, imported)
		}
		sources = append(sources, s)
		return nil
	})
	require.NoError(t, err)
	return sources
}

func isProject(module, imported string) bool {
	return imported == module || strings.HasPrefix(imported, module+"/")
}

func isStandardLibrary(module, imported string) bool {
	first, _, _ := strings.Cut(imported, "/")
	return !strings.Contains(first, ".") && !isProject(module, imported)
}

// contextOf returns the bounded context and the layer of a file under
// bounded-context/<context>/<layer>/.
func contextOf(path string) (context, layer string, ok bool) {
	parts := strings.Split(path, "/")
	if len(parts) < 4 || parts[0] != "bounded-context" {
		return "", "", false
	}
	return parts[1], parts[2], true
}

// importsAnotherContext: a bounded context imports another one.
func importsAnotherContext(module string, sources []source) []string {
	prefix := module + "/internal/bounded-context/"
	var violations []string
	for _, s := range sources {
		context, _, ok := contextOf(s.path)
		if !ok {
			continue
		}
		for _, imported := range s.imports {
			rest, isContext := strings.CutPrefix(imported, prefix)
			if !isContext {
				continue
			}
			if other, _, _ := strings.Cut(rest, "/"); other != context {
				violations = append(violations, fmt.Sprintf("%s imports %s", s.path, imported))
			}
		}
	}
	return violations
}

// contractsImportingMore: internal/contracts imports more than the standard library.
func contractsImportingMore(module string, sources []source) []string {
	var violations []string
	for _, s := range sources {
		if !strings.HasPrefix(s.path, "contracts/") {
			continue
		}
		for _, imported := range s.imports {
			if !isStandardLibrary(module, imported) {
				violations = append(violations, fmt.Sprintf("%s imports %s", s.path, imported))
			}
		}
	}
	return violations
}

// domainImportingOutward: a domain imports a package of the project, or a
// third-party package that is not allowed.
func domainImportingOutward(module string, sources []source) []string {
	var violations []string
	for _, s := range sources {
		if _, layer, ok := contextOf(s.path); !ok || layer != "domain" {
			continue
		}
		for _, imported := range s.imports {
			if isProject(module, imported) || (!isStandardLibrary(module, imported) && !thirdPartyInDomain[imported]) {
				violations = append(violations, fmt.Sprintf("%s imports %s", s.path, imported))
			}
		}
	}
	return violations
}

// applicationImportingAdapter: a use case imports an adapter.
func applicationImportingAdapter(module string, sources []source) []string {
	var violations []string
	for _, s := range sources {
		if _, layer, ok := contextOf(s.path); !ok || layer != "application" {
			continue
		}
		for _, imported := range s.imports {
			if isProject(module, imported) && strings.Contains(imported, "/adapter/") {
				violations = append(violations, fmt.Sprintf("%s imports %s", s.path, imported))
			}
		}
	}
	return violations
}

func TestTheSourcesAreFound(t *testing.T) {
	contexts := map[string]bool{}
	for _, s := range loadSources(t) {
		if context, _, ok := contextOf(s.path); ok {
			contexts[context] = true
		}
	}

	assert.True(t, contexts["annonce"], "the rules below would pass on nothing")
	assert.True(t, contexts["moderation"], "the rules below would pass on nothing")
}

func TestBoundedContextsNeverImportEachOther(t *testing.T) {
	assert.Empty(t, importsAnotherContext(modulePath(t), loadSources(t)),
		"contexts talk through the events of internal/contracts")
}

func TestContractsImportOnlyTheStandardLibrary(t *testing.T) {
	assert.Empty(t, contractsImportingMore(modulePath(t), loadSources(t)))
}

func TestDomainsImportNoProjectPackage(t *testing.T) {
	assert.Empty(t, domainImportingOutward(modulePath(t), loadSources(t)),
		"the domain imports the standard library and "+fmt.Sprint(thirdPartyInDomain))
}

func TestApplicationsNeverImportAnAdapter(t *testing.T) {
	assert.Empty(t, applicationImportingAdapter(modulePath(t), loadSources(t)))
}

func TestTheRulesSeeAViolation(t *testing.T) {
	const module = "example.com/shop"
	sources := []source{
		{path: "bounded-context/annonce/application/command/a.go", imports: []string{
			module + "/internal/bounded-context/moderation/domain",
			module + "/internal/bounded-context/annonce/adapter/driven/inmemory",
		}},
		{path: "bounded-context/annonce/domain/a.go", imports: []string{
			"time", "github.com/google/uuid", module + "/pkg/eventbus", "github.com/lib/pq",
		}},
		{path: "contracts/a.go", imports: []string{"time", module + "/pkg/eventbus"}},
	}

	assert.Equal(t, []string{
		"bounded-context/annonce/application/command/a.go imports example.com/shop/internal/bounded-context/moderation/domain",
	}, importsAnotherContext(module, sources))
	assert.Equal(t, []string{
		"bounded-context/annonce/application/command/a.go imports example.com/shop/internal/bounded-context/annonce/adapter/driven/inmemory",
	}, applicationImportingAdapter(module, sources))
	assert.Equal(t, []string{
		"bounded-context/annonce/domain/a.go imports example.com/shop/pkg/eventbus",
		"bounded-context/annonce/domain/a.go imports github.com/lib/pq",
	}, domainImportingOutward(module, sources))
	assert.Equal(t, []string{
		"contracts/a.go imports example.com/shop/pkg/eventbus",
	}, contractsImportingMore(module, sources))
}
