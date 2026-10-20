package bsdqris

import (
	"fmt"
	"strconv"
	"strings"
)

// GenerateDynamic mengonversi string QRIS Statis toko menjadi QRIS Dinamis
// dengan nominal tagihan (amount) yang terkunci secara otomatis.
//
// Aturan Arsitektur Wajib yang Dipenuhi:
//  1. Token/Node Traversal: Menghindari jebakan pencarian substring ("5802").
//  2. Preservasi Tag 62: Tag 62 bawaan acquirer (Terminal ID / Store Label) 100% utuh tanpa modifikasi.
//  3. Validasi Limit Bank Indonesia: Menolak amount <= 0 atau > Rp 10.000.000.
//  4. Checksum Integritas: Menghitung ulang CRC-16 CCITT-FALSE yang valid.
func GenerateDynamic(staticPayload string, amount int64) (string, error) {
	// 1. Validasi Batas Nominal Regulasi Bank Indonesia
	if amount < MinQRISAmount {
		return "", ErrInvalidAmount
	}
	if amount > MaxQRISAmount {
		return "", ErrExceedsQRISLimit
	}

	// 2. Parse & Validasi Integritas Payload Awal (Termasuk Checksum CRC-16)
	qrisData, err := Parse(staticPayload)
	if err != nil {
		return "", err
	}

	// 3. Format Tag 54 (Transaction Amount)
	amountStr := strconv.FormatInt(amount, 10)
	tag54 := fmt.Sprintf("54%02d%s", len(amountStr), amountStr)

	// 4. Token Traversal Reconstruction (Rebuilding Payload Tanpa Mengubah Tag Lain)
	var b strings.Builder
	tag54Inserted := false

	for _, node := range qrisData.Nodes {
		// Abaikan Tag 63 lama (akan dihitung ulang di akhir)
		if node.Tag == "63" {
			continue
		}

		// Transformasi Tag 01: Ganti "11" (Static) menjadi "12" (Dynamic)
		if node.Tag == "01" {
			b.WriteString("010212")
			continue
		}

		// Jika Tag 54 sudah ada sebelumnya, perbarui dengan nominal baru
		if node.Tag == "54" {
			b.WriteString(tag54)
			tag54Inserted = true
			continue
		}

		// Sisipkan Tag 54 tepat sebelum Tag 55 (Tip), 56, 57, atau 58 (Country Code)
		if !tag54Inserted && isInsertionTargetTag(node.Tag) {
			b.WriteString(tag54)
			tag54Inserted = true
		}

		// Seluruh tag lainnya (termasuk Tag 62, Tag 26-51, Tag 52 MCC, Tag 53 Currency)
		// ditulis apa adanya 100% utuh tanpa modifikasi
		b.WriteString(fmt.Sprintf("%s%02d%s", node.Tag, node.Length, node.Value))
	}

	// Jika payload tidak memiliki Tag 55 s/d 58, sisipkan Tag 54 sebelum Tag 63
	if !tag54Inserted {
		b.WriteString(tag54)
	}

	// 5. Tambahkan Header Tag 63 ("6304")
	b.WriteString("6304")

	// 6. Hitung Ulang Checksum CRC-16 CCITT-FALSE
	reconstructedData := b.String()
	newCRC := CalculateCRC16(reconstructedData)

	// 7. Gabungkan Buffer dengan 4 Karakter Hexadesimal Checksum Baru
	b.WriteString(newCRC)

	return b.String(), nil
}

// isInsertionTargetTag memeriksa apakah tag saat ini adalah titik injeksi Tag 54 yang sah
// (sebelum Tag 55 Tip, Tag 56 Fee Fixed, Tag 57 Fee Percentage, atau Tag 58 Country Code)
func isInsertionTargetTag(tag string) bool {
	return tag == "55" || tag == "56" || tag == "57" || tag == "58"
}
