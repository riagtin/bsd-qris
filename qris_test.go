package bsdqris_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"strconv"
	"strings"
	"testing"

	bsdqris "github.com/riagtin/bsd-qris"
)

// Helper untuk membangun QRIS statis valid untuk keperluan pengujian
func tlv(tag, value string) string {
	return tag + strings.Repeat("0", 2-len(strings.TrimSpace(strconv.Itoa(len(value))))) + strconv.Itoa(len(value)) + value
}

func createSampleStaticQRIS() string {
	tag26Val := "0016ID.CO.GOPAY.WWW" +
		"0118936000140000580219" +
		"0215000000000000001" +
		"0303UME"

	tag51Val := "0014ID.OR.GPN.WWW" +
		"0215ID1020023456789" +
		"0303UME"

	tag62Val := "0108PP-10492" +
		"0708TERM-001"

	body := "000201" +
		"010211" +
		"26" + strconv.Itoa(len(tag26Val)) + tag26Val +
		"51" + strconv.Itoa(len(tag51Val)) + tag51Val +
		"52045812" +
		"5303360" +
		"5802ID" +
		"5916WARUNG KOPI JAYA" +
		"6007BANDUNG" +
		"610540115" +
		"62" + strconv.Itoa(len(tag62Val)) + tag62Val +
		"6304"

	crc := bsdqris.CalculateCRC16(body)
	return body + crc
}

func TestCRC16Calculation(t *testing.T) {
	// Test vector standar untuk CRC-16 CCITT-FALSE
	input := "123456789"
	expected := "29B1"
	actual := bsdqris.CalculateCRC16(input)
	if actual != expected {
		t.Fatalf("CRC-16 mismatch: expected %s, got %s", expected, actual)
	}

	// Test verifikasi helper
	sampleQR := createSampleStaticQRIS()
	if !bsdqris.VerifyCRC16(sampleQR) {
		t.Fatalf("VerifyCRC16 failed for sample valid QRIS")
	}

	// Test tamper payload
	tamperedQR := sampleQR[:len(sampleQR)-1] + "0"
	if bsdqris.VerifyCRC16(tamperedQR) {
		t.Fatalf("VerifyCRC16 should have failed for tampered QRIS")
	}
}

func TestParseStaticQRIS(t *testing.T) {
	sampleQR := createSampleStaticQRIS()

	data, err := bsdqris.Parse(sampleQR)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if data.IsDynamic {
		t.Errorf("Expected IsDynamic = false, got true")
	}
	if data.MerchantName != "WARUNG KOPI JAYA" {
		t.Errorf("Expected MerchantName = WARUNG KOPI JAYA, got %s", data.MerchantName)
	}
	if data.City != "BANDUNG" {
		t.Errorf("Expected City = BANDUNG, got %s", data.City)
	}
	if data.PostalCode != "40115" {
		t.Errorf("Expected PostalCode = 40115, got %s", data.PostalCode)
	}
	if data.CurrencyCode != "360" {
		t.Errorf("Expected CurrencyCode = 360, got %s", data.CurrencyCode)
	}
	if data.AcquirerName != "GoPay" {
		t.Errorf("Expected AcquirerName = GoPay, got %s", data.AcquirerName)
	}
	if data.TerminalID != "PP-10492" && data.TerminalID != "TERM-001" {
		t.Errorf("Expected TerminalID to be parsed from Tag 62, got %s", data.TerminalID)
	}

	// Uji Tag 62 utuh di data.Tags
	expectedTag62 := "0108PP-104920708TERM-001"
	if data.Tags["62"] != expectedTag62 {
		t.Errorf("Tag 62 mismatch: expected %s, got %s", expectedTag62, data.Tags["62"])
	}
}

func TestParseCorruptedQRIS(t *testing.T) {
	sampleQR := createSampleStaticQRIS()

	// 1. Checksum rusak
	corruptedCRC := sampleQR[:len(sampleQR)-4] + "FFFF"
	_, err := bsdqris.Parse(corruptedCRC)
	if err != bsdqris.ErrInvalidCRC16 {
		t.Errorf("Expected ErrInvalidCRC16, got %v", err)
	}

	// 2. Payload terlalu pendek
	_, err = bsdqris.Parse("0002010102116304ABCD")
	if err != bsdqris.ErrPayloadTooShort {
		t.Errorf("Expected ErrPayloadTooShort, got %v", err)
	}
}

func TestGenerateDynamicQRIS(t *testing.T) {
	sampleQR := createSampleStaticQRIS()
	amount := int64(47500) // Rp 47.500

	dynamicQR, err := bsdqris.GenerateDynamic(sampleQR, amount)
	if err != nil {
		t.Fatalf("GenerateDynamic failed: %v", err)
	}

	// 1. Verifikasi Checksum QRIS Dinamis Valid
	if !bsdqris.VerifyCRC16(dynamicQR) {
		t.Fatalf("Generated dynamic QRIS has invalid CRC-16!")
	}

	// 2. Parse kembali QRIS Dinamis hasil generate
	parsed, err := bsdqris.Parse(dynamicQR)
	if err != nil {
		t.Fatalf("Failed to parse generated dynamic QRIS: %v", err)
	}

	// 3. Verifikasi Status Dynamic & Nominal
	if !parsed.IsDynamic {
		t.Errorf("Expected IsDynamic = true, got false")
	}
	if parsed.Amount != amount {
		t.Errorf("Expected Amount = %d, got %d", amount, parsed.Amount)
	}

	// 4. Verifikasi Tag 01 berubah menjadi "12"
	if parsed.Tags["01"] != "12" {
		t.Errorf("Expected Tag 01 = '12', got %s", parsed.Tags["01"])
	}

	// 5. Verifikasi Tag 54 bernilai "47500"
	if parsed.Tags["54"] != "47500" {
		t.Errorf("Expected Tag 54 = '47500', got %s", parsed.Tags["54"])
	}

	// 6. ATURAN KRUSIAL: Verifikasi Tag 62 Asli 100% Utuh Tanpa Rusak/Hilang!
	expectedTag62 := "0108PP-104920708TERM-001"
	if parsed.Tags["62"] != expectedTag62 {
		t.Fatalf("Tag 62 was corrupted! Expected %s, got %s", expectedTag62, parsed.Tags["62"])
	}

	// 7. ATURAN KRUSIAL: Verifikasi Token Traversal tidak tertipu substring "5802" pada NMID
	// Tag 26 harus tetap utuh dan memiliki panjang yang sama
	expectedTag26 := "0016ID.CO.GOPAY.WWW011893600014000058021902150000000000000010303UME"
	if parsed.Tags["26"] != expectedTag26 {
		t.Fatalf("Tag 26 was corrupted by substring search! Expected %s, got %s", expectedTag26, parsed.Tags["26"])
	}

	// Tag 52 MCC 5812 harus tetap utuh
	if parsed.Tags["52"] != "5812" {
		t.Fatalf("Tag 52 MCC was corrupted! Expected 5812, got %s", parsed.Tags["52"])
	}
}

func TestBankIndonesiaLimitValidation(t *testing.T) {
	sampleQR := createSampleStaticQRIS()

	// 1. Amount 0 atau negatif harus ditolak
	_, err := bsdqris.GenerateDynamic(sampleQR, 0)
	if err != bsdqris.ErrInvalidAmount {
		t.Errorf("Expected ErrInvalidAmount for amount 0, got %v", err)
	}

	_, err = bsdqris.GenerateDynamic(sampleQR, -5000)
	if err != bsdqris.ErrInvalidAmount {
		t.Errorf("Expected ErrInvalidAmount for negative amount, got %v", err)
	}

	// 2. Amount pas Rp 10.000.000 (Maksimal Valid)
	_, err = bsdqris.GenerateDynamic(sampleQR, 10000000)
	if err != nil {
		t.Errorf("Expected Rp 10.000.000 to be valid, got error: %v", err)
	}

	// 3. Amount Rp 10.000.001 (Melebihi Limit Bank Indonesia) harus ditolak
	_, err = bsdqris.GenerateDynamic(sampleQR, 10000001)
	if err != bsdqris.ErrExceedsQRISLimit {
		t.Errorf("Expected ErrExceedsQRISLimit for 10.000.001, got %v", err)
	}
}

func TestToPNGAndDecodeImageRoundTrip(t *testing.T) {
	sampleQR := createSampleStaticQRIS()
	dynamicQR, err := bsdqris.GenerateDynamic(sampleQR, 25000)
	if err != nil {
		t.Fatalf("GenerateDynamic failed: %v", err)
	}

	// 1. Generate PNG Bytes
	pngBytes, err := bsdqris.ToPNG(dynamicQR, 256)
	if err != nil {
		t.Fatalf("ToPNG failed: %v", err)
	}

	// Verifikasi Magic Bytes PNG (\x89PNG\r\n\x1a\n)
	if len(pngBytes) < 8 || !bytes.HasPrefix(pngBytes, []byte("\x89PNG\r\n\x1a\n")) {
		t.Fatalf("ToPNG did not return valid PNG signature")
	}

	// 2. Decode kembali dari bytes PNG
	decodedPayload, err := bsdqris.DecodeImage(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("DecodeImage failed: %v", err)
	}

	if decodedPayload != dynamicQR {
		t.Fatalf("Decoded QR payload does not match original!\nOriginal: %s\nDecoded : %s", dynamicQR, decodedPayload)
	}
}

func TestImageDecompressionBombProtection(t *testing.T) {
	// Buat gambar dummy dengan dimensi melampaui batas (misal 5000 x 100 px)
	// Ukuran file PNG kecil tapi dimensi piksel besar
	img := image.NewRGBA(image.Rect(0, 0, 5000, 100))
	for x := 0; x < 5000; x++ {
		img.Set(x, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	}

	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		t.Fatalf("Failed to encode dummy image: %v", err)
	}

	// DecodeImage harus menolak gambar berdimensi > 4096px
	_, err = bsdqris.DecodeImage(&buf)
	if err != bsdqris.ErrInvalidImageDimension {
		t.Errorf("Expected ErrInvalidImageDimension, got %v", err)
	}
}

func TestImageMaxSizeBytesProtection(t *testing.T) {
	// Buat stream dummy melebihi 5 MB
	largeReader := strings.NewReader(strings.Repeat("A", int(bsdqris.MaxUploadSizeBytes)+100))

	_, err := bsdqris.DecodeImage(largeReader)
	if err != bsdqris.ErrImageTooLarge {
		t.Errorf("Expected ErrImageTooLarge, got %v", err)
	}
}

func BenchmarkParse(b *testing.B) {
	sampleQR := createSampleStaticQRIS()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bsdqris.Parse(sampleQR)
	}
}

func BenchmarkGenerateDynamic(b *testing.B) {
	sampleQR := createSampleStaticQRIS()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = bsdqris.GenerateDynamic(sampleQR, 47500)
	}
}
