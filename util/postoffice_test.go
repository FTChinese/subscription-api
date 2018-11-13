package util

import (
	"html/template"
	"os"
	"testing"
)

const letter = `
Dear {{.Name}},
{{if .Attended}}
It was a pleasure to see you at the wedding.
{{- else}}
It is a shame you couldn't make it to the wedding.
{{- end}}
{{with .Gift -}}
Thank you for the lovely {{.}}
{{end}}
Best wishes,
Josie
`

type Recipient struct {
	Name, Gift string
	Attended   bool
}

var recipients = []Recipient{
	{"Aunt Mildred", "bone china tea set", true},
	{"Uncle John", "moleskin pants", false},
	{"Cousin Rodney", "", false},
}

func TestTemplate(t *testing.T) {
	tpl := template.Must(template.New("letter").Parse(letter))

	for _, r := range recipients {
		err := tpl.Execute(os.Stdout, r)

		if err != nil {
			t.Log("executing template:", err)
		}
	}
}
