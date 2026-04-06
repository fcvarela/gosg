package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/qmuntal/gltf"
	_ "golang.org/x/image/bmp"
)

func doToGLTF(modelPath, outputDir string) {
	data, err := os.ReadFile(modelPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", modelPath, err)
		os.Exit(1)
	}

	meshes := parseModelFile(data)
	if len(meshes) == 0 {
		fmt.Fprintln(os.Stderr, "No meshes found in model file")
		os.Exit(1)
	}

	baseName := strings.TrimSuffix(filepath.Base(modelPath), filepath.Ext(modelPath))
	if outputDir == "" {
		outputDir = "."
	}
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output dir: %v\n", err)
		os.Exit(1)
	}

	doc := gltf.NewDocument()

	rootNodeIdx := len(doc.Nodes)
	rootNode := &gltf.Node{Name: baseName}
	doc.Nodes = append(doc.Nodes, rootNode)
	doc.Scenes[0].Nodes = append(doc.Scenes[0].Nodes, rootNodeIdx)

	for i, m := range meshes {
		meshName := m.Name
		if meshName == "" {
			meshName = fmt.Sprintf("mesh_%d", i)
		}

		nodeIdx := addMeshToDoc(doc, &m, meshName)
		rootNode.Children = append(rootNode.Children, nodeIdx)
	}

	outputPath := filepath.Join(outputDir, baseName+".glb")
	if err := gltf.Save(doc, outputPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outputPath, err)
		os.Exit(1)
	}
	fmt.Printf("Converted %d mesh(es) to %s\n", len(meshes), outputPath)
}

func addMeshToDoc(doc *gltf.Document, m *meshData, name string) int {
	primitive := gltf.Primitive{
		Attributes: map[string]int{},
		Mode:       gltf.PrimitiveTriangles,
	}

	// Positions
	if len(m.Positions) > 0 {
		positions := bytesToFloat32(m.Positions)
		minPos, maxPos := computeMinMax3(positions)
		accIdx := addAccessor(doc, positions, gltf.AccessorVec3, gltf.ComponentFloat, &minPos, &maxPos)
		primitive.Attributes[gltf.POSITION] = accIdx
	}

	// Normals
	if len(m.Normals) > 0 {
		normals := bytesToFloat32(m.Normals)
		accIdx := addAccessor(doc, normals, gltf.AccessorVec3, gltf.ComponentFloat, nil, nil)
		primitive.Attributes[gltf.NORMAL] = accIdx
	}

	// Texcoords — model stores vec3, glTF expects vec2
	if len(m.Tcoords) > 0 {
		tc3 := bytesToFloat32(m.Tcoords)
		tc2 := make([]float32, (len(tc3)/3)*2)
		for i := 0; i < len(tc3)/3; i++ {
			tc2[i*2+0] = tc3[i*3+0]
			tc2[i*2+1] = tc3[i*3+1]
		}
		accIdx := addAccessor(doc, tc2, gltf.AccessorVec2, gltf.ComponentFloat, nil, nil)
		primitive.Attributes[gltf.TEXCOORD_0] = accIdx
	}

	// Indices (uint16)
	if len(m.Indices) > 0 {
		indices := bytesToUint16(m.Indices)
		accIdx := addIndexAccessor(doc, indices)
		primitive.Indices = &accIdx
	}

	// Material
	matIdx := addMaterial(doc, m)
	primitive.Material = &matIdx

	meshIdx := len(doc.Meshes)
	doc.Meshes = append(doc.Meshes, &gltf.Mesh{
		Name:       name,
		Primitives: []*gltf.Primitive{&primitive},
	})

	nodeIdx := len(doc.Nodes)
	doc.Nodes = append(doc.Nodes, &gltf.Node{
		Name: name,
		Mesh: gltf.Index(meshIdx),
	})

	return nodeIdx
}

func addMaterial(doc *gltf.Document, m *meshData) int {
	mat := &gltf.Material{
		Name: m.Name,
		PBRMetallicRoughness: &gltf.PBRMetallicRoughness{
			MetallicFactor:  f64p(1.0),
			RoughnessFactor: f64p(1.0),
		},
	}

	if m.State == "pbr-transparent" {
		mat.AlphaMode = gltf.AlphaBlend
	} else {
		mat.AlphaMode = gltf.AlphaOpaque
	}

	// Base color texture
	if len(m.AlbedoMap) > 0 {
		texIdx := addImageTexture(doc, m.AlbedoMap)
		if texIdx >= 0 {
			mat.PBRMetallicRoughness.BaseColorTexture = &gltf.TextureInfo{Index: texIdx}
		}
	}

	// Normal texture
	if len(m.NormalMap) > 0 {
		texIdx := addImageTexture(doc, m.NormalMap)
		if texIdx >= 0 {
			mat.NormalTexture = &gltf.NormalTexture{Index: gltf.Index(texIdx)}
		}
	}

	// Metallic-roughness combined texture (glTF: R=0, G=roughness, B=metallic)
	if len(m.RoughMap) > 0 || len(m.MetalMap) > 0 {
		combined := combineMetallicRoughness(m.RoughMap, m.MetalMap)
		if combined != nil {
			texIdx := addImageTexture(doc, combined)
			if texIdx >= 0 {
				mat.PBRMetallicRoughness.MetallicRoughnessTexture = &gltf.TextureInfo{Index: texIdx}
			}
		}
	}

	matIdx := len(doc.Materials)
	doc.Materials = append(doc.Materials, mat)
	return matIdx
}

func combineMetallicRoughness(roughData, metalData []byte) []byte {
	var roughImg, metalImg image.Image
	var err error

	if len(roughData) > 0 {
		roughImg, _, err = image.Decode(bytes.NewReader(roughData))
		if err != nil {
			return nil
		}
	}
	if len(metalData) > 0 {
		metalImg, _, err = image.Decode(bytes.NewReader(metalData))
		if err != nil {
			return nil
		}
	}

	var w, h int
	if roughImg != nil {
		w = roughImg.Bounds().Dx()
		h = roughImg.Bounds().Dy()
	} else if metalImg != nil {
		w = metalImg.Bounds().Dx()
		h = metalImg.Bounds().Dy()
	} else {
		return nil
	}

	var roughRGBA, metalRGBA *image.RGBA
	if roughImg != nil {
		roughRGBA = image.NewRGBA(image.Rect(0, 0, w, h))
		draw.Draw(roughRGBA, roughRGBA.Bounds(), roughImg, image.Point{}, draw.Src)
	}
	if metalImg != nil {
		metalRGBA = image.NewRGBA(image.Rect(0, 0, w, h))
		draw.Draw(metalRGBA, metalRGBA.Bounds(), metalImg, image.Point{}, draw.Src)
	}

	combined := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := (y*w + x) * 4
			var rough, metal byte
			if roughRGBA != nil {
				rough = roughRGBA.Pix[i]
			}
			if metalRGBA != nil {
				metal = metalRGBA.Pix[i]
			}
			combined.Pix[i+0] = 0
			combined.Pix[i+1] = rough
			combined.Pix[i+2] = metal
			combined.Pix[i+3] = 255
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, combined); err != nil {
		return nil
	}
	return buf.Bytes()
}

func addImageTexture(doc *gltf.Document, data []byte) int {
	if len(data) == 0 {
		return -1
	}

	mimeType := detectMimeType(data)

	bufIdx := 0
	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}
	buf := doc.Buffers[bufIdx]
	offset := int(len(buf.Data))
	buf.Data = append(buf.Data, data...)
	buf.ByteLength = int(len(buf.Data))

	bvIdx := len(doc.BufferViews)
	doc.BufferViews = append(doc.BufferViews, &gltf.BufferView{
		Buffer:     bufIdx,
		ByteOffset: offset,
		ByteLength: int(len(data)),
	})

	imgIdx := len(doc.Images)
	doc.Images = append(doc.Images, &gltf.Image{
		MimeType:   mimeType,
		BufferView: gltf.Index(bvIdx),
	})

	texIdx := len(doc.Textures)
	doc.Textures = append(doc.Textures, &gltf.Texture{
		Source: gltf.Index(imgIdx),
	})

	return texIdx
}

func addAccessor(doc *gltf.Document, data []float32, accType gltf.AccessorType, compType gltf.ComponentType, min, max *[3]float64) int {
	bufIdx := 0
	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}
	buf := doc.Buffers[bufIdx]

	for len(buf.Data)%4 != 0 {
		buf.Data = append(buf.Data, 0)
	}

	offset := int(len(buf.Data))
	rawBytes := float32SliceToBytes(data)
	buf.Data = append(buf.Data, rawBytes...)
	buf.ByteLength = int(len(buf.Data))

	bvIdx := len(doc.BufferViews)
	doc.BufferViews = append(doc.BufferViews, &gltf.BufferView{
		Buffer:     bufIdx,
		ByteOffset: offset,
		ByteLength: int(len(rawBytes)),
	})

	components := accessorComponents(accType)
	count := int(len(data)) / int(components)

	acc := &gltf.Accessor{
		BufferView:    gltf.Index(bvIdx),
		ComponentType: compType,
		Count:         count,
		Type:          accType,
	}
	if min != nil {
		acc.Min = min[:]
	}
	if max != nil {
		acc.Max = max[:]
	}

	accIdx := len(doc.Accessors)
	doc.Accessors = append(doc.Accessors, acc)
	return accIdx
}

func addIndexAccessor(doc *gltf.Document, indices []uint16) int {
	bufIdx := 0
	if len(doc.Buffers) == 0 {
		doc.Buffers = append(doc.Buffers, &gltf.Buffer{})
	}
	buf := doc.Buffers[bufIdx]

	for len(buf.Data)%4 != 0 {
		buf.Data = append(buf.Data, 0)
	}

	offset := int(len(buf.Data))
	rawBytes := uint16SliceToBytes(indices)
	buf.Data = append(buf.Data, rawBytes...)
	buf.ByteLength = int(len(buf.Data))

	bvIdx := len(doc.BufferViews)
	doc.BufferViews = append(doc.BufferViews, &gltf.BufferView{
		Buffer:     bufIdx,
		ByteOffset: offset,
		ByteLength: int(len(rawBytes)),
	})

	accIdx := len(doc.Accessors)
	doc.Accessors = append(doc.Accessors, &gltf.Accessor{
		BufferView:    gltf.Index(bvIdx),
		ComponentType: gltf.ComponentUshort,
		Count:         int(len(indices)),
		Type:          gltf.AccessorScalar,
	})
	return accIdx
}

func computeMinMax3(data []float32) ([3]float64, [3]float64) {
	mn := [3]float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64}
	mx := [3]float64{-math.MaxFloat64, -math.MaxFloat64, -math.MaxFloat64}
	for i := 0; i < len(data)/3; i++ {
		for c := 0; c < 3; c++ {
			v := float64(data[i*3+c])
			if v < mn[c] {
				mn[c] = v
			}
			if v > mx[c] {
				mx[c] = v
			}
		}
	}
	return mn, mx
}

func bytesToFloat32(b []byte) []float32 {
	data := make([]float32, len(b)/4)
	for i := range data {
		data[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4]))
	}
	return data
}

func bytesToUint16(b []byte) []uint16 {
	data := make([]uint16, len(b)/2)
	for i := range data {
		data[i] = binary.LittleEndian.Uint16(b[i*2 : (i+1)*2])
	}
	return data
}

func float32SliceToBytes(data []float32) []byte {
	buf := make([]byte, len(data)*4)
	for i, v := range data {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(v))
	}
	return buf
}

func uint16SliceToBytes(data []uint16) []byte {
	buf := make([]byte, len(data)*2)
	for i, v := range data {
		binary.LittleEndian.PutUint16(buf[i*2:], v)
	}
	return buf
}

func accessorComponents(t gltf.AccessorType) int {
	switch t {
	case gltf.AccessorScalar:
		return 1
	case gltf.AccessorVec2:
		return 2
	case gltf.AccessorVec3:
		return 3
	case gltf.AccessorVec4:
		return 4
	default:
		return 1
	}
}

func detectMimeType(data []byte) string {
	if len(data) > 3 {
		if data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' {
			return "image/png"
		}
		if data[0] == 0xFF && data[1] == 0xD8 {
			return "image/jpeg"
		}
	}
	return "application/octet-stream"
}

func f64p(v float64) *float64 {
	return &v
}
