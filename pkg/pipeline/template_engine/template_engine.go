/*
Copyright 2025 Richard Kosegi

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template_engine

import (
	"bytes"
	"strconv"
	"strings"
	"text/template"

	sprig "github.com/go-task/slim-sprig/v3"
)

var defFuncs = template.FuncMap{
	"toYaml":         toYamlFunc,
	"isEmpty":        isEmptyFunc,
	"unflatten":      unflattenFunc,
	"fileExists":     fileExistsFunc,
	"mergeFiles":     mergeFilesFunc,
	"isDir":          isDirFunc,
	"glob":           globFunc,
	"dom2yaml":       dom2yamlFunc,
	"dom2json":       dom2jsonFunc,
	"dom2properties": dom2propertiesFunc,
	"domdiff":        domDiffFunc,
	"urlParseQuery":  urlParseQuery,
	"fileGlob":       fileGlobFunc,
}

type templateEngine struct {
	fm template.FuncMap
}

func renderTemplate(tmplStr string, data interface{}, fm template.FuncMap) (string, error) {
	tmpl := template.New("tmpl").Funcs(fm)
	tmpl.Funcs(template.FuncMap{
		"tpl": tplFunc(tmpl),
	})
	_, err := tmpl.Parse(tmplStr)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	err = tmpl.Execute(&out, data)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func possiblyTemplate(in string) bool {
	openIdx := strings.Index(in, "{{")
	if openIdx == -1 {
		return false
	}
	closeIdx := strings.Index(in[openIdx:], "}}")
	return closeIdx > 0
}

func renderLenientTemplate(tmpl string, data map[string]interface{}, fm template.FuncMap) string {
	if possiblyTemplate(tmpl) {
		if val, err := renderTemplate(tmpl, data, fm); err != nil {
			return tmpl
		} else {
			return val
		}
	}
	return tmpl
}

func (te templateEngine) RenderLenient(tmpl string, data map[string]interface{}) string {
	return renderLenientTemplate(tmpl, data, te.fm)
}

func (te templateEngine) RenderSliceLenient(tmpls []string, data map[string]interface{}) []string {
	out := make([]string, len(tmpls))
	for i, tmpl := range tmpls {
		out[i] = renderLenientTemplate(tmpl, data, te.fm)
	}
	return out
}

func (te templateEngine) RenderMapLenient(input map[string]interface{}, data map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	for k, v := range input {
		if s, ok := v.(string); ok {
			ret[k] = te.RenderLenient(s, data)
			continue
		}
		if m, ok := v.(map[string]interface{}); ok {
			ret[k] = te.RenderMapLenient(m, data)
			continue
		}
		ret[k] = v
	}
	return ret
}

func (te templateEngine) Render(tmpl string, data map[string]interface{}) (string, error) {
	return renderTemplate(tmpl, data, te.fm)
}

func (te templateEngine) EvalBool(template string, data map[string]interface{}) (bool, error) {
	val, err := te.Render(template, data)
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(strings.TrimSpace(val))
}

type TemplateEngineOpt func(*templateEngine)

func AddFuncMap(fm template.FuncMap) TemplateEngineOpt {
	return func(p *templateEngine) {
		for k, v := range fm {
			p.fm[k] = v
		}
	}
}

func DefaultFuncMapOpt() TemplateEngineOpt {
	return AddFuncMap(defFuncs)
}

func NewTemplateEngine(opts ...TemplateEngineOpt) TemplateEngine {
	te := &templateEngine{fm: template.FuncMap{}}
	for _, opt := range opts {
		opt(te)
	}
	return te
}

func DefaultTemplateEngine() TemplateEngine {
	return NewTemplateEngine(
		DefaultFuncMapOpt(),
		AddFuncMap(sprig.TxtFuncMap()))
}
