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

package pipeline

import (
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestExecuteImportOp(t *testing.T) {
	var (
		is ImportOpSpec
		gd dom.ContainerBuilder
	)
	t.Run("import valid JSON", func(t *testing.T) {
		gd = dom.ContainerNode()
		is = ImportOpSpec{
			File: "../../testdata/doc1.json",
			Path: "step1.data",
			Mode: ParseFileModeJson,
		}

		assert.NoError(t, New(WithData(gd)).Execute(&is))
		assert.Equal(t, "c", gd.Lookup("step1.data.root.list1[2]").AsLeaf().Value())
	})

	t.Run("parsing YAML file as JSON should lead to error", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../testdata/doc1.yaml",
			Mode: ParseFileModeJson,
		}
		assert.Error(t, New().Execute(&is))
	})

	t.Run("import YAML into specific path", func(t *testing.T) {
		gd = dom.ContainerNode()
		is = ImportOpSpec{
			File: "../../testdata/doc1.yaml",
			Mode: ParseFileModeYaml,
			Path: "step1.data",
		}
		assert.NoError(t, New(WithData(gd)).Execute(&is))
		assert.Equal(t, 456, gd.Lookup("step1.data.level1.level2a.level3b").AsLeaf().Value())
	})

	t.Run("import text file into specific path", func(t *testing.T) {
		gd = dom.ContainerNode()
		is = ImportOpSpec{
			File: "../../testdata/doc1.yaml",
			Mode: ParseFileModeText,
			Path: "step3",
		}
		assert.NoError(t, New(WithData(gd)).Execute(&is))
		assert.NotEmpty(t, gd.Lookup("step3").AsLeaf().Value())
		assert.Contains(t, is.String(), "path=step3,mode=text")
	})

	t.Run("import YAML document as binary file file into specific path", func(t *testing.T) {
		gd = dom.ContainerNode()
		is = ImportOpSpec{
			File: "../../testdata/doc1.yaml",
			Mode: ParseFileModeBinary,
			Path: "files.doc1",
		}
		assert.NoError(t, New(WithData(gd)).Execute(&is))
		assert.NotEmpty(t, gd.Lookup("files.doc1").AsLeaf().Value())
	})

	t.Run("import JSON document in default mode", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/doc1.json",
			Path: "files.doc1_json",
		}
		assert.NoError(t, New().Execute(&is))
		assert.Contains(t, is.String(), "path=files.doc1_json,mode=")
	})

	t.Run("attempt to import non-existent file", func(t *testing.T) {
		is = ImportOpSpec{
			File: "non-existent-file.ext",
			Path: "something",
		}
		assert.Error(t, New().Execute(&is))
	})

	t.Run("import properties file into specific path", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/props1.properties",
			Mode: ParseFileModeProperties,
			Path: "props",
		}
		assert.NoError(t, New().Execute(&is))
	})

	t.Run("import text into root (empty path) should fail", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/props1.properties",
			Mode: ParseFileModeText,
		}
		assert.Error(t, New().Execute(&is))
	})

	t.Run("import JSON document directly to root (no path)", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/doc1.json",
			Mode: ParseFileModeJson,
		}
		assert.NoError(t, New().Execute(&is))
	})

	t.Run("import using invalid mode should fail", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/props1.properties",
			Path: "something",
			Mode: "invalid-mode",
		}
		assert.Error(t, New().Execute(&is))
	})

	t.Run("import valid XML/HTML", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/doc1.html",
			Mode: ParseFileModeXml,
		}
		assert.NoError(t, New().Execute(&is))
	})

	t.Run("import XML/HTML with invalid xpath should fail", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/doc1.html",
			Mode: ParseFileModeXml,
			Xml: &XmlImportOptions{
				Query: &ValOrRef{Val: "////bad"},
			},
		}
		assert.Error(t, New().Execute(&is))
	})

	t.Run("import XML/HTML with invalid layout should fail", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/doc1.html",
			Mode: ParseFileModeXml,
			Xml: &XmlImportOptions{
				Layout: ptr(XmlLayout("invalid")),
			},
		}
		assert.Error(t, New(WithData(gd)).Execute(&is))
	})

	t.Run("import XML/HTML with empty query should fallback to /html", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/doc1.html",
			Mode: ParseFileModeXml,
			Xml: &XmlImportOptions{
				Query: &ValOrRef{Ref: "non.existing"},
			},
		}
		assert.NoError(t, New().Execute(&is))
	})

	t.Run("import XML/HTML with non-resolvable query should fail", func(t *testing.T) {
		is = ImportOpSpec{
			File: "../../testdata/doc1.html",
			Mode: ParseFileModeXml,
			Xml: &XmlImportOptions{
				Query: &ValOrRef{Val: "/this/does/not/work"},
			},
		}
		assert.Error(t, New(WithData(dom.ContainerNode())).Execute(&is))
	})
}
