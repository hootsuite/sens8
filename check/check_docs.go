package check

import (
	"os"
	"strings"
	"sort"
	"text/template"
	"github.com/hootsuite/sens8/util"
)

type CheckUsage struct {
	Description string
	Flags       string
}

type CheckDocs struct {
	CheckUsage
	Id        string
	Resources []string
}

// Docs gets the usage docs for all registered checks
func Docs() []CheckDocs {
	docs := map[string]CheckDocs{}
	keys := []string{}
	for name, factory := range checkFactories {
		conf := CheckConfig{
			Name: name,
			Command: name,
			Id: name,
			Argv:[]string{name},
		}
		// ignore errors. factory might fail due to missing args
		c, _ := factory.factory(conf)
		docs[name] = CheckDocs{
			CheckUsage: c.Usage(),
			Id: name,
			Resources: factory.resourceTypes,
		}

		keys = append(keys, name)
	}

	// sort the docs so the markdown diffs etc are consistent.
	sorted := []CheckDocs{}
	sort.Strings(keys)
	for _, name := range keys {
		sorted = append(sorted, docs[name])
	}
	return sorted
}

var checkDocsMarkdownTpl = `
Checks Commands
===============

Get latest docs via: ` + "`" + `./sens8 -check-commands` + "`" + `

{{ range $doc := .Docs }}
### ` + "`" + `{{ $doc.Id }}` + "`" + `

**Resources**: {{ StringsJoin $doc.Resources ", " }}

{{ $doc.Description }}

` + "```" + `
{{ $doc.Flags }}
` + "```" + `

{{ end }}
`

// PrintCheckDocsMarkdown prints the usage docs for checks in markdown meant for external (github) docs
func PrintCheckDocsMarkdown() {
	PrintCheckDocs(checkDocsMarkdownTpl)
}

var checkDocsTextTpl = `
Checks Commands

{{ range $doc := .Docs }}
{{ $doc.Id }}
{{ HeaderLine $doc.Id }}

Resources: {{ StringsJoin $doc.Resources ", " }}

{{ $doc.Description }}

{{ $doc.Flags }}

{{ end }}
`

// PrintCheckDocsText prints the usage docs for checks in a format suitable for CLI
func PrintCheckDocsText() {
	PrintCheckDocs(checkDocsTextTpl)
}

// PrintCheckDocs prints the usage docs for checks with the given template
func PrintCheckDocs(tpl string) {
	headerLine := func(s string) string {
		return util.PadRight("", "=", len(s))
	}

	fm := template.FuncMap{
		"StringsJoin": strings.Join,
		"HeaderLine": headerLine,
	}
	t := template.Must(template.New("checkDocs").Funcs(fm).Parse(tpl))

	data := struct{ Docs []CheckDocs }{Docs()}
	err := t.Execute(os.Stdout, data)
	if err != nil {
		panic(err)
	}
}
