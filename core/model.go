package core

import (
	"encoding/binary"
	"fmt"
	"math"
	"path/filepath"

	"github.com/golang/glog"
)

// modelMesh holds the raw data for a single mesh within a model.
type modelMesh struct {
	Indices    []byte
	Positions  []byte
	Normals    []byte
	Tangents   []byte
	Bitangents []byte
	Tcoords    []byte
	AlbedoMap  []byte
	NormalMap   []byte
	RoughMap   []byte
	MetalMap   []byte
	State      string
	Name       string
}

// model holds the deserialized model data.
type model struct {
	Meshes []modelMesh
}

// parseModel decodes a protobuf-encoded model file using a minimal wire format parser.
// This avoids depending on the protobuf library for this simple schema.
func parseModel(data []byte) (*model, error) {
	m := &model{}
	offset := 0
	for offset < len(data) {
		fieldTag, n := decodeVarint(data[offset:])
		if n == 0 {
			return nil, fmt.Errorf("bad varint at offset %d", offset)
		}
		offset += n

		fieldNum := fieldTag >> 3
		wireType := fieldTag & 0x7

		if fieldNum == 1 && wireType == 2 { // repeated Mesh (field 1, length-delimited)
			length, n := decodeVarint(data[offset:])
			if n == 0 {
				return nil, fmt.Errorf("bad length varint at offset %d", offset)
			}
			offset += n

			meshData := data[offset : offset+int(length)]
			offset += int(length)

			mesh, err := parseMesh(meshData)
			if err != nil {
				return nil, err
			}
			m.Meshes = append(m.Meshes, mesh)
		} else {
			// Skip unknown field
			offset = skipField(data, offset, wireType)
			if offset < 0 {
				return nil, fmt.Errorf("failed to skip field %d", fieldNum)
			}
		}
	}
	return m, nil
}

func parseMesh(data []byte) (modelMesh, error) {
	var m modelMesh
	offset := 0
	for offset < len(data) {
		fieldTag, n := decodeVarint(data[offset:])
		if n == 0 {
			break
		}
		offset += n

		fieldNum := fieldTag >> 3
		wireType := fieldTag & 0x7

		if wireType == 2 { // length-delimited (bytes or string)
			length, n := decodeVarint(data[offset:])
			if n == 0 {
				return m, fmt.Errorf("bad length varint")
			}
			offset += n
			value := data[offset : offset+int(length)]
			offset += int(length)

			switch fieldNum {
			case 1:
				m.Indices = value
			case 2:
				m.Positions = value
			case 3:
				m.Normals = value
			case 4:
				m.Tangents = value
			case 5:
				m.Bitangents = value
			case 6:
				m.Tcoords = value
			case 7:
				m.AlbedoMap = value
			case 8:
				m.NormalMap = value
			case 9:
				m.RoughMap = value
			case 10:
				m.MetalMap = value
			case 11:
				m.State = string(value)
			case 12:
				m.Name = string(value)
			}
		} else {
			offset = skipField(data, offset, wireType)
			if offset < 0 {
				return m, fmt.Errorf("failed to skip field %d", fieldNum)
			}
		}
	}
	return m, nil
}

func decodeVarint(buf []byte) (uint64, int) {
	var x uint64
	var s uint
	for i, b := range buf {
		if i >= 10 {
			return 0, 0
		}
		if b < 0x80 {
			return x | uint64(b)<<s, i + 1
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
	return 0, 0
}

func skipField(data []byte, offset int, wireType uint64) int {
	switch wireType {
	case 0: // varint
		for offset < len(data) {
			if data[offset]&0x80 == 0 {
				return offset + 1
			}
			offset++
		}
		return -1
	case 1: // 64-bit
		return offset + 8
	case 2: // length-delimited
		length, n := decodeVarint(data[offset:])
		if n == 0 {
			return -1
		}
		return offset + n + int(length)
	case 5: // 32-bit
		return offset + 4
	default:
		return -1
	}
}

// LoadModel parses model data from a raw resource and returns a node.
func LoadModel(name string, res []byte) *Node {
	m, err := parseModel(res)
	if err != nil {
		glog.Fatalf("Failed to parse model %s: %v", name, err)
	}

	basename := filepath.Base(name)
	parentNode := NewNode(basename)
	for i := range m.Meshes {
		node := NewNode(basename + fmt.Sprintf("-%d", i))
		node.state = resourceManager.State(m.Meshes[i].State)

		textureDescriptor := TextureDescriptor{
			Mipmaps:  true,
			Filter:   TextureFilterMipmapLinear,
			WrapMode: TextureWrapModeRepeat,
		}

		if len(m.Meshes[i].AlbedoMap) > 0 {
			node.MaterialData().SetTexture("albedoTex", renderer.NewTextureFromImageData(m.Meshes[i].AlbedoMap, textureDescriptor))
		}
		if len(m.Meshes[i].NormalMap) > 0 {
			node.MaterialData().SetTexture("normalTex", renderer.NewTextureFromImageData(m.Meshes[i].NormalMap, textureDescriptor))
		}
		if len(m.Meshes[i].RoughMap) > 0 {
			node.MaterialData().SetTexture("roughTex", renderer.NewTextureFromImageData(m.Meshes[i].RoughMap, textureDescriptor))
		}
		if len(m.Meshes[i].MetalMap) > 0 {
			node.MaterialData().SetTexture("metalTex", renderer.NewTextureFromImageData(m.Meshes[i].MetalMap, textureDescriptor))
		}

		mesh := renderer.NewMesh()
		mesh.SetName(node.name)
		mesh.SetPositions(bytesToFloat(m.Meshes[i].Positions))
		mesh.SetNormals(bytesToFloat(m.Meshes[i].Normals))
		mesh.SetTextureCoordinates(bytesToFloat(m.Meshes[i].Tcoords))
		mesh.SetIndices(bytesToShort(m.Meshes[i].Indices))
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
