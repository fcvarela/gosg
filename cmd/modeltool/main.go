package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "pack":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: modeltool pack <manifest.yaml>")
			os.Exit(1)
		}
		doPack(os.Args[2])
	case "unpack":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: modeltool unpack <model.model> [output_dir]")
			os.Exit(1)
		}
		outputDir := ""
		if len(os.Args) >= 4 {
			outputDir = os.Args[3]
		}
		doUnpack(os.Args[2], outputDir)
	case "togltf":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "Usage: modeltool togltf <file.model> [output_dir]")
			os.Exit(1)
		}
		outputDir := ""
		if len(os.Args) >= 4 {
			outputDir = os.Args[3]
		}
		doToGLTF(os.Args[2], outputDir)
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `modeltool — pack/unpack gosg .model files

Usage:
  modeltool pack <manifest.yaml>     Pack meshes into a .model file
  modeltool unpack <file.model> [dir] Unpack a .model file into binary files
  modeltool togltf <file.model> [dir] Convert a .model file to .glb (glTF Binary)

Manifest format (YAML):
  output: mymodel.model
  meshes:
    - state: pbr-opaque
      name: fuselage
      indices: mesh0/indices.bin
      positions: mesh0/positions.bin
      normals: mesh0/normals.bin
      tcoords: mesh0/tcoords.bin
      albedo: mesh0/albedo.png
      normal: mesh0/normal.png
      rough: mesh0/rough.png
      metal: mesh0/metal.png
    - state: pbr-transparent
      name: canopy
      indices: mesh1/indices.bin
      positions: mesh1/positions.bin
      normals: mesh1/normals.bin
      tcoords: mesh1/tcoords.bin
      albedo: mesh1/albedo.png

File paths in the manifest are relative to the manifest file's directory.`)
}

// --- Manifest ---

type manifest struct {
	Output string         `yaml:"output"`
	Meshes []meshManifest `yaml:"meshes"`
}

type meshManifest struct {
	State      string `yaml:"state"`
	Name       string `yaml:"name"`
	Indices    string `yaml:"indices"`
	Positions  string `yaml:"positions"`
	Normals    string `yaml:"normals"`
	Tangents   string `yaml:"tangents"`
	Bitangents string `yaml:"bitangents"`
	Tcoords    string `yaml:"tcoords"`
	Albedo     string `yaml:"albedo"`
	Normal     string `yaml:"normal"`
	Rough      string `yaml:"rough"`
	Metal      string `yaml:"metal"`
}

// --- Protobuf wire format encoding ---

func encodeVarint(v uint64) []byte {
	var buf [10]byte
	n := 0
	for v >= 0x80 {
		buf[n] = byte(v) | 0x80
		v >>= 7
		n++
	}
	buf[n] = byte(v)
	return buf[:n+1]
}

func encodeBytes(fieldNum uint64, data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	var buf []byte
	buf = append(buf, encodeVarint(fieldNum<<3|2)...)
	buf = append(buf, encodeVarint(uint64(len(data)))...)
	buf = append(buf, data...)
	return buf
}

func encodeString(fieldNum uint64, s string) []byte {
	if s == "" {
		return nil
	}
	return encodeBytes(fieldNum, []byte(s))
}

func encodeMesh(m meshManifest, baseDir string) []byte {
	read := func(path string) []byte {
		if path == "" {
			return nil
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(baseDir, path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", path, err)
			os.Exit(1)
		}
		return data
	}

	var mesh []byte
	mesh = append(mesh, encodeBytes(1, read(m.Indices))...)
	mesh = append(mesh, encodeBytes(2, read(m.Positions))...)
	mesh = append(mesh, encodeBytes(3, read(m.Normals))...)
	mesh = append(mesh, encodeBytes(4, read(m.Tangents))...)
	mesh = append(mesh, encodeBytes(5, read(m.Bitangents))...)
	mesh = append(mesh, encodeBytes(6, read(m.Tcoords))...)
	mesh = append(mesh, encodeBytes(7, read(m.Albedo))...)
	mesh = append(mesh, encodeBytes(8, read(m.Normal))...)
	mesh = append(mesh, encodeBytes(9, read(m.Rough))...)
	mesh = append(mesh, encodeBytes(10, read(m.Metal))...)
	mesh = append(mesh, encodeString(11, m.State)...)
	mesh = append(mesh, encodeString(12, m.Name)...)
	return mesh
}

// --- Pack ---

func doPack(manifestPath string) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading manifest: %v\n", err)
		os.Exit(1)
	}

	var mf manifest
	if err := yaml.Unmarshal(data, &mf); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing manifest: %v\n", err)
		os.Exit(1)
	}

	if mf.Output == "" {
		fmt.Fprintln(os.Stderr, "Error: manifest must specify 'output'")
		os.Exit(1)
	}

	baseDir := filepath.Dir(manifestPath)
	outputPath := mf.Output
	if !filepath.IsAbs(outputPath) {
		outputPath = filepath.Join(baseDir, outputPath)
	}

	var model []byte
	for _, m := range mf.Meshes {
		meshData := encodeMesh(m, baseDir)
		model = append(model, encodeBytes(1, meshData)...)
	}

	if err := os.WriteFile(outputPath, model, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", outputPath, err)
		os.Exit(1)
	}
	fmt.Printf("Packed %d mesh(es) into %s (%d bytes)\n", len(mf.Meshes), outputPath, len(model))
}

// --- Protobuf wire format decoding ---

func decodeVarintBuf(buf []byte) (uint64, int) {
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
	case 0:
		for offset < len(data) {
			if data[offset]&0x80 == 0 {
				return offset + 1
			}
			offset++
		}
		return -1
	case 1:
		return offset + 8
	case 2:
		length, n := decodeVarintBuf(data[offset:])
		if n == 0 {
			return -1
		}
		return offset + n + int(length)
	case 5:
		return offset + 4
	default:
		return -1
	}
}

type meshData struct {
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

func parseMeshData(data []byte) meshData {
	var m meshData
	offset := 0
	for offset < len(data) {
		tag, n := decodeVarintBuf(data[offset:])
		if n == 0 {
			break
		}
		offset += n
		fieldNum := tag >> 3
		wireType := tag & 0x7

		if wireType == 2 {
			length, n := decodeVarintBuf(data[offset:])
			if n == 0 {
				break
			}
			offset += n
			value := make([]byte, length)
			copy(value, data[offset:offset+int(length)])
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
				break
			}
		}
	}
	return m
}

func parseModelFile(data []byte) []meshData {
	var meshes []meshData
	offset := 0
	for offset < len(data) {
		tag, n := decodeVarintBuf(data[offset:])
		if n == 0 {
			break
		}
		offset += n
		fieldNum := tag >> 3
		wireType := tag & 0x7

		if fieldNum == 1 && wireType == 2 {
			length, n := decodeVarintBuf(data[offset:])
			if n == 0 {
				break
			}
			offset += n
			meshBytes := data[offset : offset+int(length)]
			offset += int(length)
			meshes = append(meshes, parseMeshData(meshBytes))
		} else {
			offset = skipField(data, offset, wireType)
			if offset < 0 {
				break
			}
		}
	}
	return meshes
}

// --- Unpack ---

func doUnpack(input, outputDir string) {
	if outputDir == "" {
		outputDir = strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))
	}

	data, err := os.ReadFile(input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", input, err)
		os.Exit(1)
	}

	meshes := parseModelFile(data)
	fmt.Printf("Model %s: %d mesh(es)\n\n", input, len(meshes))

	// Generate a manifest for easy repacking
	mf := manifest{Output: filepath.Base(input)}

	for i, m := range meshes {
		meshDir := fmt.Sprintf("mesh_%d", i)
		fullDir := filepath.Join(outputDir, meshDir)
		if err := os.MkdirAll(fullDir, 0755); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", fullDir, err)
			os.Exit(1)
		}

		mm := meshManifest{State: m.State, Name: m.Name}

		writeField := func(name string, data []byte) string {
			if len(data) == 0 {
				return ""
			}
			relPath := filepath.Join(meshDir, name)
			path := filepath.Join(outputDir, relPath)
			if err := os.WriteFile(path, data, 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", path, err)
				os.Exit(1)
			}
			return relPath
		}

		imageExt := func(data []byte) string {
			if len(data) > 3 {
				if data[0] == 0xFF && data[1] == 0xD8 {
					return ".jpg"
				}
				if data[0] == 0x89 && data[1] == 'P' && data[2] == 'N' {
					return ".png"
				}
				if data[0] == 'B' && data[1] == 'M' {
					return ".bmp"
				}
			}
			return ".bin"
		}

		fmt.Printf("Mesh %d: state=%q name=%q  verts=%d tris=%d\n",
			i, m.State, m.Name, len(m.Positions)/12, len(m.Indices)/6)

		mm.Indices = writeField("indices.bin", m.Indices)
		mm.Positions = writeField("positions.bin", m.Positions)
		mm.Normals = writeField("normals.bin", m.Normals)
		mm.Tangents = writeField("tangents.bin", m.Tangents)
		mm.Bitangents = writeField("bitangents.bin", m.Bitangents)
		mm.Tcoords = writeField("tcoords.bin", m.Tcoords)
		mm.Albedo = writeField("albedo"+imageExt(m.AlbedoMap), m.AlbedoMap)
		mm.Normal = writeField("normal"+imageExt(m.NormalMap), m.NormalMap)
		mm.Rough = writeField("rough"+imageExt(m.RoughMap), m.RoughMap)
		mm.Metal = writeField("metal"+imageExt(m.MetalMap), m.MetalMap)

		mf.Meshes = append(mf.Meshes, mm)
	}

	// Write manifest for repacking
	manifestPath := filepath.Join(outputDir, "manifest.yaml")
	mfData, _ := yaml.Marshal(&mf)
	os.WriteFile(manifestPath, mfData, 0644)
	fmt.Printf("\nManifest written to %s\n", manifestPath)
}

var _ = binary.LittleEndian
