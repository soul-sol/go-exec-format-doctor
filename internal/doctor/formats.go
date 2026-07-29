package doctor

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"slices"
	"strings"
)

func inspectELF(header []byte) metadata {
	if len(header) < 20 {
		return metadata{format: FormatELF, targetOS: "linux", detail: "ELF header is truncated before its machine field"}
	}

	var byteOrder binary.ByteOrder = binary.LittleEndian
	if header[5] == 2 {
		byteOrder = binary.BigEndian
	}
	machine := byteOrder.Uint16(header[18:20])
	architectures := architectureForELF(machine)
	detail := "ELF executable header parsed"
	if len(architectures) == 0 {
		detail = fmt.Sprintf("ELF machine 0x%x is not mapped to a Go architecture", machine)
	}

	return metadata{
		format:        FormatELF,
		targetOS:      "linux",
		architectures: architectures,
		detail:        detail,
	}
}

func inspectMachO(header []byte) metadata {
	if len(header) < 8 {
		return metadata{format: FormatMachO, targetOS: "darwin", detail: "Mach-O header is truncated before its CPU field"}
	}

	byteOrder := machOByteOrder(header)
	cpu := byteOrder.Uint32(header[4:8])

	return metadata{
		format:        FormatMachO,
		targetOS:      "darwin",
		architectures: architectureForMachO(cpu),
		detail:        "Mach-O executable header parsed",
	}
}

func inspectMachOFat(header []byte) metadata {
	if len(header) < 8 {
		return metadata{format: FormatMachOFat, targetOS: "darwin", detail: "universal Mach-O header is truncated"}
	}

	byteOrder := machOFatByteOrder(header)
	count := int(byteOrder.Uint32(header[4:8]))
	entrySize := 20
	if bytes.Equal(header[:4], []byte{0xca, 0xfe, 0xba, 0xbf}) ||
		bytes.Equal(header[:4], []byte{0xbf, 0xba, 0xfe, 0xca}) {
		entrySize = 24
	}
	if count < 1 || count > 64 || len(header) < 8+(count*entrySize) {
		return metadata{format: FormatMachOFat, targetOS: "darwin", detail: "universal Mach-O architecture table is invalid or truncated"}
	}

	architectures := make([]string, 0, count)
	for index := range count {
		offset := 8 + (index * entrySize)
		for _, arch := range architectureForMachO(byteOrder.Uint32(header[offset : offset+4])) {
			if !slices.Contains(architectures, arch) {
				architectures = append(architectures, arch)
			}
		}
	}

	return metadata{
		format:        FormatMachOFat,
		targetOS:      "darwin",
		architectures: architectures,
		detail:        "universal Mach-O architecture table parsed",
	}
}

func inspectPE(header []byte) metadata {
	if len(header) < 0x40 {
		return metadata{format: FormatPE, targetOS: "windows", detail: "PE DOS header is truncated"}
	}

	offset := int(binary.LittleEndian.Uint32(header[0x3c:0x40]))
	if offset < 0 || offset > len(header)-6 || !bytes.Equal(header[offset:offset+4], []byte{'P', 'E', 0, 0}) {
		return metadata{format: FormatPE, targetOS: "windows", detail: "PE signature or machine field is invalid or truncated"}
	}

	return metadata{
		format:        FormatPE,
		targetOS:      "windows",
		architectures: architectureForPE(binary.LittleEndian.Uint16(header[offset+4 : offset+6])),
		detail:        "PE executable header parsed",
	}
}

func inspectScript(header []byte) metadata {
	firstLine := header
	if newline := bytes.IndexByte(header, '\n'); newline >= 0 {
		firstLine = header[:newline]
	}
	interpreter := strings.TrimSpace(string(bytes.TrimPrefix(firstLine, []byte("#!"))))
	detail := "script requires its shebang interpreter on this host"
	if interpreter == "" {
		detail = "script shebang is empty"
	}

	return metadata{
		format:      FormatScript,
		targetOS:    "unix",
		interpreter: interpreter,
		detail:      detail,
	}
}

func architectureForELF(machine uint16) []string {
	switch machine {
	case 0x03:
		return []string{"386"}
	case 0x28:
		return []string{"arm"}
	case 0x3e:
		return []string{"amd64"}
	case 0xb7:
		return []string{"arm64"}
	case 0xf3:
		return []string{"riscv64"}
	default:
		return []string{}
	}
}

func architectureForMachO(cpu uint32) []string {
	switch cpu {
	case 0x00000007:
		return []string{"386"}
	case 0x01000007:
		return []string{"amd64"}
	case 0x0000000c:
		return []string{"arm"}
	case 0x0100000c:
		return []string{"arm64"}
	default:
		return []string{}
	}
}

func architectureForPE(machine uint16) []string {
	switch machine {
	case 0x014c:
		return []string{"386"}
	case 0x8664:
		return []string{"amd64"}
	case 0x01c4:
		return []string{"arm"}
	case 0xaa64:
		return []string{"arm64"}
	default:
		return []string{}
	}
}

func hasMachOMagic(header []byte) bool {
	if len(header) < 4 {
		return false
	}

	return bytes.Equal(header[:4], []byte{0xfe, 0xed, 0xfa, 0xce}) ||
		bytes.Equal(header[:4], []byte{0xce, 0xfa, 0xed, 0xfe}) ||
		bytes.Equal(header[:4], []byte{0xfe, 0xed, 0xfa, 0xcf}) ||
		bytes.Equal(header[:4], []byte{0xcf, 0xfa, 0xed, 0xfe})
}

func hasMachOFatMagic(header []byte) bool {
	if len(header) < 4 {
		return false
	}

	return bytes.Equal(header[:4], []byte{0xca, 0xfe, 0xba, 0xbe}) ||
		bytes.Equal(header[:4], []byte{0xbe, 0xba, 0xfe, 0xca}) ||
		bytes.Equal(header[:4], []byte{0xca, 0xfe, 0xba, 0xbf}) ||
		bytes.Equal(header[:4], []byte{0xbf, 0xba, 0xfe, 0xca})
}

func machOByteOrder(header []byte) binary.ByteOrder {
	if bytes.Equal(header[:4], []byte{0xce, 0xfa, 0xed, 0xfe}) ||
		bytes.Equal(header[:4], []byte{0xcf, 0xfa, 0xed, 0xfe}) {
		return binary.LittleEndian
	}

	return binary.BigEndian
}

func machOFatByteOrder(header []byte) binary.ByteOrder {
	if bytes.Equal(header[:4], []byte{0xbe, 0xba, 0xfe, 0xca}) ||
		bytes.Equal(header[:4], []byte{0xbf, 0xba, 0xfe, 0xca}) {
		return binary.LittleEndian
	}

	return binary.BigEndian
}
