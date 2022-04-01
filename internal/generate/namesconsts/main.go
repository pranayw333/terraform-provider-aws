//go:build generate
// +build generate

package main

import (
	"bytes"
	"encoding/csv"
	"go/format"
	"log"
	"os"
	"text/template"
)

const filename = `consts_gen.go`

type ServiceDatum struct {
	ProviderNameUpper string
	ProviderPackage   string
}

type TemplateData struct {
	Services []ServiceDatum
}

const (
	// column indices of CSV
	//awsCLIV2Command         = 0
	//awsCLIV2CommandNoDashes = 1
	//goV1Package             = 2
	//goV2Package             = 3
	//providerPackageActual   = 4
	//providerPackageCorrect  = 5
	//aliases                 = 8
	//goV1ClientName          = 9
	//humanFriendly           = 10
	//brand                   = 11
	//note                    = 12
	//deprecatedEnvVar        = 14
	//envVar                  = 15
	providerPackageBoth = 6
	providerNameUpper   = 7
	exclude             = 13
)

func main() {
	f, err := os.Open("names_data.csv")
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	csvReader := csv.NewReader(f)

	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	td := TemplateData{}

	for i, l := range data {
		if i > 0 { // no header
			if l[exclude] != "" || l[providerPackageBoth] == "" {
				continue
			}

			td.Services = append(td.Services, ServiceDatum{
				ProviderNameUpper: l[providerNameUpper],
				ProviderPackage:   l[providerPackageBoth],
			})
		}
	}

	writeTemplate(tmpl, "consts", td)
}

func writeTemplate(body string, templateName string, td TemplateData) {
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("error opening file (%s): %s", filename, err)
	}

	tplate, err := template.New(templateName).Parse(body)
	if err != nil {
		log.Fatalf("error parsing template: %s", err)
	}

	var buffer bytes.Buffer
	err = tplate.Execute(&buffer, td)
	if err != nil {
		log.Fatalf("error executing template: %s", err)
	}

	contents, err := format.Source(buffer.Bytes())
	if err != nil {
		log.Fatalf("error formatting generated file: %s", err)
	}

	if _, err := f.Write(contents); err != nil {
		f.Close()
		log.Fatalf("error writing to file (%s): %s", filename, err)
	}

	if err := f.Close(); err != nil {
		log.Fatalf("error closing file (%s): %s", filename, err)
	}
}

var tmpl = `
// Code generated by internal/generate/namesconsts/main.go; DO NOT EDIT.
package names

const (
{{- range .Services }}
	{{ .ProviderNameUpper }} = "{{ .ProviderPackage }}"
{{- end }}
)
`
