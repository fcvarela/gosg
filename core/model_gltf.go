package core

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"path/filepath"

	"github.com/go-gl/mathgl/mgl64"
	"github.com/golang/glog"
	"github.com/qmuntal/gltf"
)

// LoadGLTF loads a glTF/GLB file and returns a node tree.
func LoadGLTF(name string, resourceSystem ResourceSystem) *Node {
	path := resourceSystem.ModelPath(name)
	doc, err := gltf.Open(path)
	if err != nil {
		glog.Fatalf("Failed to open glTF %s: %v", name, err)
	}

	basename := filepath.Base(name)
	root := NewNode(basename)

	if len(doc.Scenes) == 0 {
		return root
	}

	sceneIdx := 0
	if doc.Scene != nil {
		sceneIdx = *doc.Scene
	}
	scene := doc.Scenes[sceneIdx]

	for _, nodeIdx := range scene.Nodes {
		child := loadGLTFNode(doc, nodeIdx, basename)
		root.AddChild(child)
	}

	return root
}

func loadGLTFNode(doc *gltf.Document, nodeIdx int, prefix string) *Node {
	gn := doc.Nodes[nodeIdx]
	name := gn.Name
	if name == "" {
		name = fmt.Sprintf("%s-node%d", prefix, nodeIdx)
	}
	node := NewNode(name)

	// Apply transform
	if gn.Matrix != [16]float64{} && gn.Matrix != [16]float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1} {
		node.transform = mgl64.Mat4(gn.Matrix)
		node.setDirtyTransform()
		node.setDirtyBounds()
	} else {
		if gn.Translation != [3]float64{} {
			node.Translate(mgl64.Vec3(gn.Translation))
		}
		if gn.Rotation != [4]float64{0, 0, 0, 1} {
			q := mgl64.Quat{W: gn.Rotation[3], V: mgl64.Vec3{gn.Rotation[0], gn.Rotation[1], gn.Rotation[2]}}
			node.transform = node.transform.Mul4(q.Normalize().Mat4())
			node.setDirtyTransform()
			node.setDirtyBounds()
		}
		if gn.Scale != [3]float64{1, 1, 1} {
			node.Scale(mgl64.Vec3(gn.Scale))
		}
	}

	// Load mesh
	if gn.Mesh != nil {
		gm := doc.Meshes[*gn.Mesh]
		for pi := range gm.Primitives {
			primNode := node
			if len(gm.Primitives) > 1 {
				primNode = NewNode(fmt.Sprintf("%s-prim%d", name, pi))
				node.AddChild(primNode)
			}
			loadGLTFPrimitive(doc, gm.Primitives[pi], primNode)
		}
	}

	// Recurse children
	for _, childIdx := range gn.Children {
		child := loadGLTFNode(doc, childIdx, prefix)
		node.AddChild(child)
	}

	return node
}

func loadGLTFPrimitive(doc *gltf.Document, prim *gltf.Primitive, node *Node) {
	mesh := NewMesh()
	mesh.SetName(node.Name())
	mesh.SetPrimitiveType(PrimitiveTypeTriangles)

	// Positions (required)
	if posIdx, ok := prim.Attributes[gltf.POSITION]; ok {
		positions := readAccessorFloat32(doc, posIdx)
		mesh.SetPositions(positions)
	}

	// Normals
	if normIdx, ok := prim.Attributes[gltf.NORMAL]; ok {
		normals := readAccessorFloat32(doc, normIdx)
		mesh.SetNormals(normals)
	} else {
		if posIdx, ok := prim.Attributes[gltf.POSITION]; ok {
			positions := readAccessorFloat32(doc, posIdx)
			normals := generateFlatNormals(positions)
			mesh.SetNormals(normals)
		}
	}

	// Texcoords — engine expects vec3, glTF provides vec2; pad with zero Z
	if tcIdx, ok := prim.Attributes[gltf.TEXCOORD_0]; ok {
		tc2 := readAccessorFloat32(doc, tcIdx)
		acc := doc.Accessors[tcIdx]
		if acc.Type == gltf.AccessorVec2 {
			tc3 := make([]float32, (len(tc2)/2)*3)
			for i := 0; i < len(tc2)/2; i++ {
				tc3[i*3+0] = tc2[i*2+0]
				tc3[i*3+1] = tc2[i*2+1]
				tc3[i*3+2] = 0
			}
			mesh.SetTextureCoordinates(tc3)
		} else {
			mesh.SetTextureCoordinates(tc2)
		}
	} else {
		if posIdx, ok := prim.Attributes[gltf.POSITION]; ok {
			vertCount := int(doc.Accessors[posIdx].Count)
			tc := make([]float32, vertCount*3)
			mesh.SetTextureCoordinates(tc)
		}
	}

	// Indices
	if prim.Indices != nil {
		acc := doc.Accessors[*prim.Indices]
		if acc.ComponentType == gltf.ComponentUint || acc.ComponentType == gltf.ComponentFloat {
			indices := readAccessorUint32(doc, *prim.Indices)
			mesh.SetIndices32(indices)
		} else {
			indices := readAccessorUint16(doc, *prim.Indices)
			mesh.SetIndices(indices)
		}
	}

	node.SetMesh(mesh)

	// Material
	pipelineName := "pbr-opaque"
	if prim.Material != nil {
		mat := doc.Materials[*prim.Material]
		if mat.AlphaMode == gltf.AlphaBlend {
			pipelineName = "pbr-transparent"
		}

		texDesc := TextureDescriptor{
			Mipmaps:  true,
			Filter:   TextureFilterMipmapLinear,
			WrapMode: TextureWrapModeRepeat,
		}

		if mat.PBRMetallicRoughness != nil {
			pbr := mat.PBRMetallicRoughness

			if pbr.BaseColorTexture != nil {
				tex := loadGLTFTexture(doc, pbr.BaseColorTexture.Index, texDesc)
				if tex != nil {
					node.Material().SetTexture("albedoTex", tex)
				}
			}

			if pbr.MetallicRoughnessTexture != nil {
				roughTex, metalTex := splitMetallicRoughnessTexture(doc, pbr.MetallicRoughnessTexture.Index, texDesc)
				if roughTex != nil {
					node.Material().SetTexture("roughTex", roughTex)
				}
				if metalTex != nil {
					node.Material().SetTexture("metalTex", metalTex)
				}
			}
		}

		if mat.NormalTexture != nil {
			tex := loadGLTFTexture(doc, *mat.NormalTexture.Index, texDesc)
			if tex != nil {
				node.Material().SetTexture("normalTex", tex)
			}
		}
	}

	node.pipeline = resourceManager.Pipeline(pipelineName)
}

func loadGLTFTexture(doc *gltf.Document, texIdx int, desc TextureDescriptor) *Texture {
	if texIdx >= len(doc.Textures) {
		return nil
	}
	gtex := doc.Textures[texIdx]
	if gtex.Source == nil {
		return nil
	}
	img := doc.Images[*gtex.Source]

	var data []byte
	if img.BufferView != nil {
		bv := doc.BufferViews[*img.BufferView]
		buf := doc.Buffers[bv.Buffer]
		data = buf.Data[bv.ByteOffset : bv.ByteOffset+bv.ByteLength]
	} else if img.URI != "" {
		glog.Warningf("glTF: external image URI %q not supported yet", img.URI)
		return nil
	}

	if len(data) == 0 {
		return nil
	}

	return renderer.NewTextureFromImageData(data, desc)
}

func splitMetallicRoughnessTexture(doc *gltf.Document, texIdx int, desc TextureDescriptor) (*Texture, *Texture) {
	if texIdx >= len(doc.Textures) {
		return nil, nil
	}
	gtex := doc.Textures[texIdx]
	if gtex.Source == nil {
		return nil, nil
	}
	img := doc.Images[*gtex.Source]

	var data []byte
	if img.BufferView != nil {
		bv := doc.BufferViews[*img.BufferView]
		buf := doc.Buffers[bv.Buffer]
		data = buf.Data[bv.ByteOffset : bv.ByteOffset+bv.ByteLength]
	}

	if len(data) == 0 {
		return nil, nil
	}

	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		glog.Warningf("glTF: failed to decode metallicRoughness texture: %v", err)
		return nil, nil
	}

	rgba := image.NewRGBA(decoded.Bounds())
	draw.Draw(rgba, rgba.Bounds(), decoded, image.Point{}, draw.Src)

	w := uint32(rgba.Rect.Size().X)
	h := uint32(rgba.Rect.Size().Y)
	pixelCount := int(w) * int(h)

	roughPixels := make([]byte, pixelCount*4)
	metalPixels := make([]byte, pixelCount*4)
	for i := 0; i < pixelCount; i++ {
		g := rgba.Pix[i*4+1]
		b := rgba.Pix[i*4+2]
		roughPixels[i*4+0] = g
		roughPixels[i*4+1] = g
		roughPixels[i*4+2] = g
		roughPixels[i*4+3] = 255
		metalPixels[i*4+0] = b
		metalPixels[i*4+1] = b
		metalPixels[i*4+2] = b
		metalPixels[i*4+3] = 255
	}

	texDesc := TextureDescriptor{
		Width: w, Height: h, Target: TextureTarget2D,
		Format: TextureFormatRGBA, SizedFormat: TextureSizedFormatRGBA8,
		ComponentType: TextureComponentTypeUNSIGNEDBYTE,
		Mipmaps: desc.Mipmaps, Filter: desc.Filter, WrapMode: desc.WrapMode,
	}

	roughTex := renderer.NewTexture(texDesc, roughPixels)
	metalTex := renderer.NewTexture(texDesc, metalPixels)
	return roughTex, metalTex
}

func readAccessorFloat32(doc *gltf.Document, accIdx int) []float32 {
	acc := doc.Accessors[accIdx]
	bv := doc.BufferViews[*acc.BufferView]
	buf := doc.Buffers[bv.Buffer]

	componentCount := accessorTypeComponents(acc.Type)
	totalFloats := int(acc.Count) * componentCount

	byteOffset := int(bv.ByteOffset) + int(acc.ByteOffset)
	stride := int(bv.ByteStride)
	if stride == 0 {
		stride = componentCount * 4
	}

	result := make([]float32, totalFloats)
	for i := 0; i < int(acc.Count); i++ {
		base := byteOffset + i*stride
		for c := 0; c < componentCount; c++ {
			bits := binary.LittleEndian.Uint32(buf.Data[base+c*4 : base+c*4+4])
			result[i*componentCount+c] = math.Float32frombits(bits)
		}
	}
	return result
}

func readAccessorUint16(doc *gltf.Document, accIdx int) []uint16 {
	acc := doc.Accessors[accIdx]
	bv := doc.BufferViews[*acc.BufferView]
	buf := doc.Buffers[bv.Buffer]

	byteOffset := int(bv.ByteOffset) + int(acc.ByteOffset)
	stride := int(bv.ByteStride)

	count := int(acc.Count)
	result := make([]uint16, count)

	switch acc.ComponentType {
	case gltf.ComponentUbyte:
		if stride == 0 {
			stride = 1
		}
		for i := 0; i < count; i++ {
			result[i] = uint16(buf.Data[byteOffset+i*stride])
		}
	case gltf.ComponentUshort:
		if stride == 0 {
			stride = 2
		}
		for i := 0; i < count; i++ {
			result[i] = binary.LittleEndian.Uint16(buf.Data[byteOffset+i*stride : byteOffset+i*stride+2])
		}
	default:
		glog.Warningf("glTF: unsupported index component type %d for uint16 read", acc.ComponentType)
	}
	return result
}

func readAccessorUint32(doc *gltf.Document, accIdx int) []uint32 {
	acc := doc.Accessors[accIdx]
	bv := doc.BufferViews[*acc.BufferView]
	buf := doc.Buffers[bv.Buffer]

	byteOffset := int(bv.ByteOffset) + int(acc.ByteOffset)
	stride := int(bv.ByteStride)
	if stride == 0 {
		stride = 4
	}

	count := int(acc.Count)
	result := make([]uint32, count)
	for i := 0; i < count; i++ {
		result[i] = binary.LittleEndian.Uint32(buf.Data[byteOffset+i*stride : byteOffset+i*stride+4])
	}
	return result
}

func accessorTypeComponents(t gltf.AccessorType) int {
	switch t {
	case gltf.AccessorScalar:
		return 1
	case gltf.AccessorVec2:
		return 2
	case gltf.AccessorVec3:
		return 3
	case gltf.AccessorVec4:
		return 4
	case gltf.AccessorMat2:
		return 4
	case gltf.AccessorMat3:
		return 9
	case gltf.AccessorMat4:
		return 16
	default:
		return 1
	}
}

func generateFlatNormals(positions []float32) []float32 {
	normals := make([]float32, len(positions))
	vertCount := len(positions) / 3
	triCount := vertCount / 3
	for t := 0; t < triCount; t++ {
		i := t * 9
		v0 := mgl64.Vec3{float64(positions[i]), float64(positions[i+1]), float64(positions[i+2])}
		v1 := mgl64.Vec3{float64(positions[i+3]), float64(positions[i+4]), float64(positions[i+5])}
		v2 := mgl64.Vec3{float64(positions[i+6]), float64(positions[i+7]), float64(positions[i+8])}
		edge1 := v1.Sub(v0)
		edge2 := v2.Sub(v0)
		n := edge1.Cross(edge2).Normalize()
		for v := 0; v < 3; v++ {
			normals[i+v*3+0] = float32(n[0])
			normals[i+v*3+1] = float32(n[1])
			normals[i+v*3+2] = float32(n[2])
		}
	}
	return normals
}
