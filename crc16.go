package bsdqris

import (
	"fmt"
	"strings"
)

// CalculateCRC16 menghitung checksum CRC-16 CCITT-FALSE standar Bank Indonesia / ASPI.
// Parameter spesifikasi:
//   - Polynomial  : 0x1021 (x^16 + x^12 + x^5 + 1)
//   - Initial     : 0xFFFF
//   - Reflect In  : false
//   - Reflect Out : false
//   - XOR Out     : 0x0000
//   - Format      : 4 karakter Hexadesimal Huruf Kapital (contoh: "9B2F")
func CalculateCRC16(data string) string {
	crc := uint16(0xFFFF)
	polynomial := uint16(0x1021)

	for i := 0; i < len(data); i++ {
		crc ^= uint16(data[i]) << 8
		for j := 0; j < 8; j++ {
			if (crc & 0x8000) != 0 {
				crc = (crc << 1) ^ polynomial
			} else {
				crc <<= 1
			}
		}
	}

	return fmt.Sprintf("%04X", crc)
}

// VerifyCRC16 memvalidasi apakah string QRIS memiliki akhiran Tag 63 (6304XXXX)
// dan nilai checksum-nya cocok 100% dengan hasil kalkulasi CRC-16 CCITT-FALSE.
func VerifyCRC16(payload string) bool {
	payload = strings.TrimSpace(payload)
	if len(payload) < 8 {
		return false
	}

	// Tag 63 harus berada di 8 karakter terakhir: "6304XXXX"
	tag63HeaderIdx := len(payload) - 8
	if payload[tag63HeaderIdx:tag63HeaderIdx+4] != "6304" {
		return false
	}

	expectedCRC := strings.ToUpper(payload[len(payload)-4:])
	dataToHash := payload[:len(payload)-4]
	calculatedCRC := CalculateCRC16(dataToHash)

	return expectedCRC == calculatedCRC
}
