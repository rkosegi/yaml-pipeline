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

package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/gookit/color"
	"github.com/kaptinlin/jsonschema"
	"github.com/rkosegi/yaml-pipeline/schemas"
	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/rkosegi/yaml-toolkit/fluent"
	"github.com/rkosegi/yaml-toolkit/path"
	"github.com/rkosegi/yaml-toolkit/props"
	"github.com/xlab/treeprint"
)

var pp = props.NewPathParser()

// ApplyVarsToDom takes map of key-to-any and puts them into DOM container under given prefix, ie "vars".
func ApplyVarsToDom(kvs map[string]interface{}, prefix string, gd dom.ContainerBuilder) {
	if kvs == nil {
		return
	}
	p := pp.MustParse(prefix)
	for k, v := range kvs {
		gd.Set(path.ChildOf(p, path.Simple(k)), dom.LeafNode(v))
	}
}

// ApplyValues takes slice of k=v strings and put them into DOM container.
func ApplyValues(gd dom.ContainerBuilder, vals []string) {
	for _, val := range vals {
		parts := strings.SplitAfterN(val, "=", 2)
		key := strings.Replace(parts[0], "=", "", 1)
		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		p := pp.MustParse(key)
		gd.Set(p, dom.LeafNode(value))
	}
}

func DumpSchemaEvalResultToTree(parent treeprint.Tree, details []*jsonschema.EvaluationResult) {
	for _, d := range details {
		x := parent.AddBranch(fmt.Sprintf("%s => %s",
			color.FgLightBlue.Render(d.InstanceLocation),
			color.Green.Render(d.EvaluationPath)))
		if !d.IsValid() {
			for k, v := range d.Errors {
				x.AddBranch(color.Red.Render(fmt.Sprintf("ERR: %s: %v", k, v)))
			}
		}
		if len(d.Details) > 0 {
			DumpSchemaEvalResultToTree(x, d.Details)
		}
	}
}

func ValidateFileAgainstSchema(file string) (*jsonschema.EvaluationResult, error) {
	compiler := jsonschema.NewCompiler()
	schemaBytes := schemas.PipelineV1Schema
	schema, err := compiler.Compile(schemaBytes)
	if err != nil {
		return nil, err
	}
	df, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer func(df *os.File) {
		_ = df.Close()
	}(df)

	doc := make(map[string]interface{})
	if err = fluent.DefaultFileDecoderProvider(file)(df, &doc); err != nil {
		return nil, err
	}
	res := schema.Validate(doc)
	return res, nil
}

func GetLogTag(v ...interface{}) (string, bool) {
	if len(v) < 2 {
		return "", false
	}
	if tag, isString := v[0].(string); isString {
		if strings.HasPrefix(tag, "tag::") {
			return tag[5:], true
		}
	}
	return "", false
}
