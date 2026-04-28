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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rkosegi/yaml-toolkit/dom"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestValOrRefMarshalYaml(t *testing.T) {
	t.Run("Marshal ref", func(t *testing.T) {
		var (
			out  bytes.Buffer
			node yaml.Node
		)
		assert.NoError(t, yaml.NewEncoder(&out).Encode(&ValOrRef{Ref: "path.to.elem"}))
		assert.NoError(t, yaml.NewDecoder(&out).Decode(&node))
		assert.Equal(t, yaml.MappingNode, node.Content[0].Kind)
	})
	t.Run("Marshal direct value", func(t *testing.T) {
		var (
			out  bytes.Buffer
			node yaml.Node
		)
		assert.NoError(t, yaml.NewEncoder(&out).Encode(&ValOrRef{Val: "abc"}))
		assert.NoError(t, yaml.NewDecoder(&out).Decode(&node))
		assert.Equal(t, yaml.ScalarNode, node.Content[0].Kind)
	})
}

func TestValOrRefUnmarshalYaml(t *testing.T) {
	t.Run("unmarshal embedded", func(t *testing.T) {
		var buf bytes.Buffer
		vIn := OsOpSpec{Stat: &OsOpStatSpec{Path: &ValOrRef{Val: "/tmp/1.txt"}}}
		assert.NoError(t, yaml.NewEncoder(&buf).Encode(vIn))
		var vOut OsOpSpec
		assert.NoError(t, yaml.NewDecoder(&buf).Decode(&vOut))
		assert.Equal(t, "/tmp/1.txt", vOut.Stat.Path.Val)
	})

	t.Run("Unmarshal empty val+ref empty should fail", func(t *testing.T) {
		y1 := `
ref: ""
val: ""
`
		var v ValOrRef
		assert.Error(t, yaml.NewDecoder(strings.NewReader(y1)).Decode(&v))
	})

}

func TestSourceOfData(t *testing.T) {
	var (
		o   dom.Leaf
		err error
		sd  *SourceOfData
	)
	gd := dom.ContainerNode().AddValue("test", dom.LeafNode("abc"))
	ac := newMockActBuilder().data(gd).build()

	t.Run("tree", func(t *testing.T) {
		sd = &SourceOfData{
			Tree: &ValOrRef{Ref: "test", isRef: true},
		}
		o, err = sd.Resolve(ac)
		assert.NoError(t, err)
		assert.Equal(t, "abc", o.Value())
	})
	t.Run("file", func(t *testing.T) {
		t.Run("valid", func(t *testing.T) {
			fpath := filepath.Join(t.TempDir(), "file.txt")
			assert.NoError(t, os.WriteFile(fpath, []byte("123XYZ"), 0640))
			sd = &SourceOfData{
				File: &ValOrRef{Val: fpath},
			}
			o, err = sd.Resolve(ac)
			assert.NoError(t, err)
			assert.Equal(t, "123XYZ", o.Value())
		})
		t.Run("non existent", func(t *testing.T) {
			sd = &SourceOfData{
				File: &ValOrRef{Val: "/this/should/not/exist"},
			}
			_, err = sd.Resolve(ac)
			assert.Error(t, err)
		})
	})
	t.Run("none of valid options provided", func(t *testing.T) {
		sd = &SourceOfData{}
		_, err = sd.Resolve(ac)
		assert.Error(t, err)
	})
}

func TestEncodeDataAs(t *testing.T) {
	var (
		data []byte
		err  error
	)
	gd := dom.ContainerNode().
		AddValue("test", dom.LeafNode("abc")).
		AddValue("test2", dom.LeafNode([]byte("Ola"))).
		AddValue("test3", dom.LeafNode(6.3)).
		AddValue("list", dom.ListNode(
			dom.LeafNode(1),
			dom.LeafNode(2),
		))
	t.Run("unknown", func(t *testing.T) {
		x := EncodeDataAs("unknown")
		_, err = x.Encode(gd)
		assert.Error(t, err)
	})
	t.Run("json", func(t *testing.T) {
		t.Run("leaf - string", func(t *testing.T) {
			data, err = EncodeDataAsJson.Encode(gd.Get(pp.MustParse("test")))
			assert.NoError(t, err)
			assert.Equal(t, `"abc"`+"\n", string(data))
		})
		t.Run("leaf - float", func(t *testing.T) {
			data, err = EncodeDataAsJson.Encode(gd.Get(pp.MustParse("test3")))
			assert.NoError(t, err)
			assert.Equal(t, "6.3\n", string(data))
		})
	})
	t.Run("yaml", func(t *testing.T) {
		t.Run("leaf - string", func(t *testing.T) {
			data, err = EncodeDataAsYaml.Encode(gd.Get(pp.MustParse("test")))
			assert.NoError(t, err)
			assert.Equal(t, "abc\n", string(data))
		})
	})
	t.Run("binary", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			data, err = EncodeDataAsBinary.Encode(nil)
			assert.NoError(t, err)
			assert.Nil(t, data)
		})
		t.Run("unsupported - container", func(t *testing.T) {
			_, err = EncodeDataAsBinary.Encode(gd)
			assert.Error(t, err)
		})
		t.Run("leaf - string", func(t *testing.T) {
			data, err = EncodeDataAsBinary.Encode(gd.Get(pp.MustParse("test")))
			assert.NoError(t, err)
			assert.Equal(t, "abc", string(data))
		})
		t.Run("leaf - byte slice", func(t *testing.T) {
			data, err = EncodeDataAsBinary.Encode(gd.Get(pp.MustParse("test2")))
			assert.NoError(t, err)
			assert.Equal(t, "Ola", string(data))
		})
		t.Run("unsupported leaf - float", func(t *testing.T) {
			_, err = EncodeDataAsBinary.Encode(gd.Get(pp.MustParse("test3")))
			assert.Error(t, err)
		})
	})

}
