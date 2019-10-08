package checks

import (
	"bytes"
	"sort"
	"strings"
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
			Name:    name,
			Command: name,
			Id:      name,
			Argv:    []string{name},
		}
		// ignore errors. factory might fail due to missing args
		c, _ := factory.factory(conf)
		docs[name] = CheckDocs{
			CheckUsage: c.Usage(),
			Id:         name,
			Resources:  factory.resourceTypes,
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

var CheckDocsMarkdownTpl = `
Checks Command Documentation
============================

Get latest docs via: ` + "`" + `./sens8 -check-docs` + "`" + `

{{ range $doc := .Docs }}
### ` + "`" + `{{ $doc.Id }}` + "`" + `

**Resources**: {{ StringsJoin $doc.Resources ", " }}

{{ $doc.Description }}

` + "```" + `
{{ $doc.Flags }}
` + "```" + `

{{ end }}
`

// GenCheckDocsMarkdown generates the usage docs for checks in markdown meant for external (github) docs
func GenCheckDocsMarkdown() string {
	return GenCheckDocs(CheckDocsMarkdownTpl)
}

var CheckDocsTextTpl = `
Checks Command Documentation

{{ range $doc := .Docs }}
{{ $doc.Id }}
{{ HeaderLine $doc.Id }}

Resources: {{ StringsJoin $doc.Resources ", " }}

{{ $doc.Description }}

{{ $doc.Flags }}

{{ end }}
`

// GenCheckDocsText generates the usage docs for checks in a format suitable for CLI
func GenCheckDocsText() string {
	return GenCheckDocs(CheckDocsTextTpl)
}

// GenCheckDocs generates the usage docs for checks with the given template
func GenCheckDocs(tpl string) string {
	headerLine := func(s string) string {
		return util.PadRight("", "=", len(s))
	}

	fm := template.FuncMap{
		"StringsJoin": strings.Join,
		"HeaderLine":  headerLine,
	}
	t := template.Must(template.New("checkDocs").Funcs(fm).Parse(tpl))

	data := struct{ Docs []CheckDocs }{Docs()}
	var out bytes.Buffer
	err := t.Execute(&out, data)
	if err != nil {
		panic(err)
	}
	return out.String()
}
