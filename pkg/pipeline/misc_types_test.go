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

func TestFileWriter(t *testing.T) {
	var (
		data []byte
		err  error
	)
	t.Run("create new file", func(t *testing.T) {
		fpath := filepath.Join(t.TempDir(), "file.txt")
		fw := &fileWriter{fpath: fpath}
		assert.NoError(t, fw.init())
		_, err = fw.Write([]byte("test"))
		assert.NoError(t, err)
		assert.NoError(t, fw.Close())
		assert.NoError(t, fw.Close())
		data, err = os.ReadFile(fpath)
		assert.NoError(t, err)
		assert.Equal(t, []byte("test"), data)
	})
	t.Run("append to existing file", func(t *testing.T) {
		fpath := filepath.Join(t.TempDir(), "file.txt")
		assert.NoError(t, os.WriteFile(fpath, []byte("1234"), 0640))
		fw := &fileWriter{fpath: fpath, append: true}
		assert.NoError(t, fw.init())
		_, err = fw.Write([]byte("56"))
		assert.NoError(t, err)
		assert.NoError(t, fw.Close())
		data, err = os.ReadFile(fpath)
		assert.NoError(t, err)
		assert.Equal(t, []byte("123456"), data)
	})
	t.Run("init not called", func(t *testing.T) {
		fpath := filepath.Join(t.TempDir(), "file.txt")
		fw := &fileWriter{fpath: fpath}
		_, err = fw.Write([]byte(nil))
		assert.Error(t, err)
	})
}

func TestDomLeafNodeWriter(t *testing.T) {
	var (
		err error
		gd  dom.ContainerBuilder
		nw  *domLeafNodeWriter
	)
	leafPath := pp.MustParse("test")
	t.Run("default", func(t *testing.T) {

		gd = dom.ContainerNode()
		nw = &domLeafNodeWriter{leafPath: leafPath, rootNode: gd, buf: new(bytes.Buffer)}
		_, err = nw.Write([]byte("Ola"))
		assert.NoError(t, err)
		assert.NoError(t, nw.Close())
		assert.Equal(t, "Ola", gd.Get(leafPath).AsLeaf().Value())
	})
	t.Run("append", func(t *testing.T) {
		gd = dom.ContainerNode().Set(leafPath, dom.LeafNode("1+1"))
		nw = &domLeafNodeWriter{leafPath: leafPath, rootNode: gd, buf: new(bytes.Buffer), append: true}
		_, err = nw.Write([]byte("=2"))
		assert.NoError(t, err)
		assert.NoError(t, nw.Close())
		assert.Equal(t, "1+1=2", gd.Get(leafPath).AsLeaf().Value())
	})
}

func TestStreamDestination(t *testing.T) {
	var (
		sd   *StreamDestination
		data []byte
		err  error
	)
	t.Run("empty", func(t *testing.T) {
		sd = &StreamDestination{}
		assert.NoError(t, sd.Resolve(newMockActBuilder().build()))
		_, err = sd.Write([]byte{})
		assert.NoError(t, err)
		assert.NoError(t, sd.Close())
	})
	t.Run("uninitialized", func(t *testing.T) {
		sd = &StreamDestination{}
		_, err = sd.Write([]byte{})
		assert.Error(t, err)
	})
	t.Run("tree", func(t *testing.T) {
		gd := dom.ContainerNode()
		sd = &StreamDestination{Tree: &ValOrRef{Val: "test"}}
		assert.NoError(t, sd.Resolve(newMockActBuilder().data(gd).build()))
		_, err = sd.Write([]byte("Hello"))
		assert.NoError(t, sd.Close())
		assert.NoError(t, err)
		assert.Equal(t, "Hello", gd.Child("test").AsLeaf().Value())
	})
	t.Run("file", func(t *testing.T) {
		fpath := filepath.Join(t.TempDir(), "file.txt")
		sd = &StreamDestination{File: &ValOrRef{Val: fpath}}
		assert.NoError(t, sd.Resolve(newMockActBuilder().build()))
		_, err = sd.Write([]byte("abcd"))
		assert.NoError(t, err)
		assert.NoError(t, sd.Close())
		data, err = os.ReadFile(fpath)
		assert.Equal(t, []byte("abcd"), data)
	})

	t.Run("clone", func(t *testing.T) {
		fpath := filepath.Join(t.TempDir(), "myFile.txt")
		gd := dom.ContainerNode().AddValue("forEach", dom.LeafNode(fpath))
		sd = &StreamDestination{File: &ValOrRef{Ref: "forEach", isRef: true}}
		ac := newMockActBuilder().data(gd).build()
		sd = safeCloneStreamDestination(sd, ac)
		assert.NoError(t, sd.Resolve(ac))
		_, err = sd.Write([]byte("Ola"))
		assert.NoError(t, err)
		assert.NoError(t, sd.Close())
		data, err = os.ReadFile(fpath)
		assert.Equal(t, []byte("Ola"), data)
		assert.NoError(t, err)

	})
}
