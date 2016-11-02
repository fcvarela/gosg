package core

import (
	"fmt"
	"testing"
)

var (
	expectedStringHashes map[string]uint32
	expectedU32Hashes    map[uint32]uint32
)

func init() {
	expectedStringHashes := make(map[string]uint32)
	expectedU32Hashes := make(map[uint32]uint32)

	expectedU32Hashes[1] = 307143837
	expectedU32Hashes[2] = 614320443
	expectedU32Hashes[3] = 920874438
	expectedU32Hashes[4] = 1228640886
	expectedU32Hashes[5] = 1534473963
	expectedU32Hashes[6] = 1841781645
	expectedU32Hashes[7] = 2148040719
	expectedU32Hashes[8] = 2457281772
	expectedU32Hashes[9] = 2762295624
	expectedU32Hashes[10] = 3068980695

	expectedStringHashes["universe:1:galaxy:1"] = 2390584335
	expectedStringHashes["universe:1:galaxy:2"] = 2128471459
	expectedStringHashes["universe:1:galaxy:3"] = 7363022
	expectedStringHashes["universe:1:galaxy:4"] = 4291432279
	expectedStringHashes["universe:1:galaxy:5"] = 2131316235
	expectedStringHashes["universe:1:galaxy:6"] = 2999750580
	expectedStringHashes["universe:1:galaxy:7"] = 2693663858
	expectedStringHashes["universe:1:galaxy:8"] = 4201988196
	expectedStringHashes["universe:1:galaxy:9"] = 849830729
	expectedStringHashes["universe:1:galaxy:10"] = 1623302695

}

func TestStringHashes(t *testing.T) {
	for key, refHash := range expectedStringHashes {
		testHash := HashStringToU32(key)
		if testHash != refHash {
			t.Error("String hash mismatch")
		}
	}
}

func BenchmarkHashU32(b *testing.B) {
	for i := uint32(1); i < 1E4; i++ {
		HashU32(i)
	}
}

func TestU32Hashes(t *testing.T) {
	for key, refHash := range expectedU32Hashes {
		testHash := HashU32(key)
		if testHash != refHash {
			t.Error("U32 hash mismatch")
		}
	}
}

func BenchmarkHashStringToU32(b *testing.B) {
	for i := uint32(1); i < 1E4; i++ {
		HashStringToU32(fmt.Sprintf("SomeObject:%d", i))
	}
}
