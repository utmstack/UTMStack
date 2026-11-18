package bytecode

import (
	"encoding/binary"
	"runtime"
	"strconv"
)

// versionCompareAPI follows ClamAV 1.5.4's version_compare bytecode host
// function. Its numeric accumulators deliberately persist across numeric
// fields; that is observable when a later field follows an equal one. The
// reference uses unsigned long, which is 32 bits on Windows and 64 bits on
// supported 64-bit Unix hosts. Invalid VM pointers are rejected by invoke
// before this function is called.
func versionCompareAPI(args []uint64, mem *memory, bytesWork func(uint64) error) (uint64, error) {
	leftLen, rightLen := int(uint32(args[1])), int(uint32(args[3]))
	if err := bytesWork(uint64(leftLen) + uint64(rightLen)); err != nil {
		return 0, err
	}
	left, err := mem.read(args[0], leftLen)
	if err != nil {
		return 0, err
	}
	right, err := mem.read(args[2], rightLen)
	if err != nil {
		return 0, err
	}
	mask := ^uint64(0)
	if runtime.GOOS == "windows" || strconv.IntSize == 32 {
		mask = 0xffffffff
	}
	i, j := 0, 0
	var leftNumber, rightNumber uint64
	for {
		for i < len(left) && j < len(right) && left[i] == right[j] &&
			!asciiDigit(left[i]) && !asciiDigit(right[j]) {
			i++
			j++
		}
		if i == len(left) && j == len(right) {
			return 0, nil
		}
		if i == len(left) {
			return apiFailure, nil
		}
		if j == len(right) {
			return 1, nil
		}
		if !asciiDigit(left[i]) || !asciiDigit(right[j]) {
			if left[i] < right[j] {
				return apiFailure, nil
			}
			return 1, nil
		}
		for i < len(left) && asciiDigit(left[i]) {
			leftNumber = (10*leftNumber + uint64(left[i]-'0')) & mask
			i++
		}
		for j < len(right) && asciiDigit(right[j]) {
			rightNumber = (10*rightNumber + uint64(right[j]-'0')) & mask
			j++
		}
		if leftNumber < rightNumber {
			return apiFailure, nil
		}
		if leftNumber > rightNumber {
			return 1, nil
		}
	}
}

func asciiDigit(b byte) bool { return b >= '0' && b <= '9' }

// checkPlatformAPI compares the three platform words in the same field
// widths and order as the reference. A full-field mask is the ANY wildcard.
func checkPlatformAPI(args []uint64) uint64 {
	env := environmentRecord()
	value := [3]uint32{
		binary.LittleEndian.Uint32(env[0:]),
		binary.LittleEndian.Uint32(env[4:]),
		binary.LittleEndian.Uint32(env[8:]),
	}
	for word, fields := range [3][]struct{ shift, mask uint32 }{
		{{24, 0xff}, {20, 0xf}, {16, 0xf}, {8, 0xff}, {0, 0xff}},
		{{28, 0xf}, {24, 0xf}, {16, 0xff}, {8, 0xff}, {0, 0xff}},
		{{24, 0xff}, {16, 0xff}, {8, 0xff}, {0, 0xff}},
	} {
		query := uint32(args[word])
		for _, field := range fields {
			q := (query >> field.shift) & field.mask
			v := (value[word] >> field.shift) & field.mask
			if q != field.mask && q != v {
				return 0
			}
		}
	}
	return 1
}
