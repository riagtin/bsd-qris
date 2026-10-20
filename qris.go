// Package bsdqris adalah modul bersama resmi PT. BINTANG STUDIO DIGITAL
// untuk pemrosesan, validasi, dan konversi QRIS Standar Nasional (ASPI / Bank Indonesia)
// berbasis spesifikasi EMVCo Merchant-Presented Mode (MPM).
//
// Modul ini digunakan oleh ekosistem Trio Flagship SaaS Enterprise:
//   - PENGPOS (Cloud POS & Inventory OS)
//   - OmniFix (Workshop Management & After-Sales OS)
//   - BSDWeb (Platform Pusat & Showcase)
package bsdqris

// NewDynamicQRIS adalah helper pintas untuk mem-parse QRIS statis,
// mengonversinya menjadi dinamis dengan amount, dan mengembalikan string payload final.
func NewDynamicQRIS(staticPayload string, amount int64) (string, error) {
	return GenerateDynamic(staticPayload, amount)
}

// NewDynamicQRISWithImage adalah helper pintas untuk menghasilkan string dinamis
// beserta byte gambar PNG dalam satu pemanggilan.
func NewDynamicQRISWithImage(staticPayload string, amount int64, imageSize int) (string, []byte, error) {
	dynamicPayload, err := GenerateDynamic(staticPayload, amount)
	if err != nil {
		return "", nil, err
	}

	pngBytes, err := ToPNG(dynamicPayload, imageSize)
	if err != nil {
		return "", nil, err
	}

	return dynamicPayload, pngBytes, nil
}
