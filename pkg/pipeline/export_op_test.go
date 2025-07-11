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
	"os"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
)

func TestExportOpDo(t *testing.T) {
	var (
		eo  *ExportOp
		err error
	)
	f, err := os.CreateTemp("", "yt_export*.json")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	removeFilesLater(t, f)
	t.Logf("created temporary file: %s", f.Name())
	eo = &ExportOp{
		File:   &ValOrRef{Val: f.Name()},
		Path:   &ValOrRef{Val: "root.sub1"},
		Format: OutputFormatJson,
	}
	assert.Contains(t, eo.String(), f.Name())
	d := dom.ContainerNode()
	d.AddValueAt("root.sub1.sub2", dom.LeafNode(123))
	err = eo.Do(newMockActBuilder().data(d).build())
	assert.NoError(t, err)
	fi, err := os.Stat(f.Name())
	assert.NotNil(t, fi)
	assert.NoError(t, err)

	eo = &ExportOp{
		File:   &ValOrRef{Val: f.Name()},
		Path:   &ValOrRef{Val: "root.sub1.sub2"},
		Format: OutputFormatText,
	}
	err = eo.Do(newMockActBuilder().data(d).build())
	assert.NoError(t, err)

	eo = &ExportOp{
		File:   &ValOrRef{Val: f.Name()},
		Path:   &ValOrRef{Val: "root.sub1"},
		Format: OutputFormatText,
	}
	err = eo.Do(newMockActBuilder().data(d).build())
	assert.Error(t, err)
}

func TestExportOpDoInvalidDirectory(t *testing.T) {
	eo := &ExportOp{
		File:   &ValOrRef{Val: "/invalid/dir/file.yaml"},
		Format: OutputFormatYaml,
	}
	assert.Error(t, eo.Do(mockEmptyActCtx()))
}

func TestExportOpDoInvalidOutFormat(t *testing.T) {
	eo := &ExportOp{
		Format: "invalid-format",
	}
	assert.Error(t, eo.Do(mockEmptyActCtx()))
}

func TestExportOpDoNonExistentPath(t *testing.T) {
	f, err := os.CreateTemp("", "yt_export*.json")
	assert.NoError(t, err)
	if err != nil {
		return
	}
	removeFilesLater(t, f)
	eo := &ExportOp{
		File:   &ValOrRef{Val: f.Name()},
		Path:   &ValOrRef{Val: "this.Path.does.not.exist"},
		Format: OutputFormatProperties,
	}
	assert.NoError(t, eo.Do(mockEmptyActCtx()))
}

func TestExportOpCloneWith(t *testing.T) {
	eo := &ExportOp{
		File:   &ValOrRef{Val: "/tmp/out.{{ .Format }}"},
		Path:   &ValOrRef{Val: "root.sub10.{{ .Sub }}"},
		Format: "{{ .Format }}",
	}
	d := dom.ContainerNode()
	d.AddValueAt("Format", dom.LeafNode("yaml"))
	d.AddValueAt("Sub", dom.LeafNode("sub20"))
	eo = eo.CloneWith(newMockActBuilder().data(d).build()).(*ExportOp)
	assert.Equal(t, "root.sub10.sub20", eo.Path.Val)
	assert.Equal(t, OutputFormatYaml, eo.Format)
	assert.Equal(t, "/tmp/out.yaml", eo.File.Val)
}
