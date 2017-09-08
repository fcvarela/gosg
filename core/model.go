package core

import (
	"encoding/binary"
	"fmt"
	"math"
	"path/filepath"

	"github.com/fcvarela/gosg/protos"
	"github.com/golang/glog"
	"github.com/golang/protobuf/proto"
)

// LoadModel parses model data from a raw resource and returns a node ready
// to insert into the screnegraph
func LoadModel(name string, res []byte) *Node {
	var model = &protos.Model{}
	if err := proto.Unmarshal(res, model); err != nil {
		glog.Fatalln(err)
	}

	basename := filepath.Base(name)
	parentNode := NewNode(basename)
	for i := 0; i < len(model.Meshes); i++ {
		node := NewNode(basename + fmt.Sprintf("-%d", i))
		node.state = resourceManager.State(model.Meshes[i].State)

		// get textures
		textureDescriptor := TextureDescriptor{
			Mipmaps:  true,
			Filter:   TextureFilterMipmapLinear,
			WrapMode: TextureWrapModeRepeat,
		}

		if len(model.Meshes[i].AlbedoMap) > 0 {
			var albedoTexture = renderSystem.NewTextureFromImageData(model.Meshes[i].AlbedoMap, textureDescriptor)
			node.MaterialData().SetTexture("albedoTex", albedoTexture)
		}

		if len(model.Meshes[i].NormalMap) > 0 {
			var normalTexture = renderSystem.NewTextureFromImageData(model.Meshes[i].NormalMap, textureDescriptor)
			node.MaterialData().SetTexture("normalTex", normalTexture)
		}

		if len(model.Meshes[i].RoughMap) > 0 {
			var roughTexture = renderSystem.NewTextureFromImageData(model.Meshes[i].RoughMap, textureDescriptor)
			node.MaterialData().SetTexture("roughTex", roughTexture)
		}

		if len(model.Meshes[i].MetalMap) > 0 {
			var metalTexture = renderSystem.NewTextureFromImageData(model.Meshes[i].MetalMap, textureDescriptor)
			node.MaterialData().SetTexture("metalTex", metalTexture)
		}

		// set mesh data
		mesh := renderSystem.NewMesh()
		mesh.SetName(node.name)
		mesh.SetPositions(bytesToFloat(model.Meshes[i].Positions))
		mesh.SetNormals(bytesToFloat(model.Meshes[i].Normals))
		mesh.SetTextureCoordinates(bytesToFloat(model.Meshes[i].Tcoords))
		mesh.SetIndices(bytesToShort(model.Meshes[i].Indices))
		mesh.SetPrimitiveType(PrimitiveTypeTriangles)

		node.SetMesh(mesh)
		parentNode.AddChild(node)
	}

	return parentNode
}

func bytesToFloat(b []byte) []float32 {
	data := make([]float32, len(b)/4)
	for i := range data {
		data[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4]))
	}
	return data
}

func bytesToShort(b []byte) []uint16 {
	data := make([]uint16, len(b)/2)
	for i := range data {
		data[i] = binary.LittleEndian.Uint16(b[i*2 : (i+1)*2])
	}
	return data
}
