package bsdqris

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/makiuchi-d/gozxing"
	gozxingqr "github.com/makiuchi-d/gozxing/qrcode"
	skip2qr "github.com/skip2/go-qrcode"
)

// ToPNG menghasilkan raw bytes gambar PNG dari string payload QRIS.
// Cocok digunakan untuk melampirkan QR ke nota servis PDF (Maroto v2)
// atau pesan WhatsApp gateway (whatsmeow) pada OmniFix.
// Jika size <= 0, default ukuran adalah 256 piksel.
func ToPNG(payload string, size int) ([]byte, error) {
	if size <= 0 {
		size = 256
	}

	// Recovery Level Medium (15% data restore) standar untuk QRIS
	pngBytes, err := skip2qr.Encode(payload, skip2qr.Medium, size)
	if err != nil {
		return nil, err
	}

	return pngBytes, nil
}

// DecodeImage mengekstrak string teks QRIS dari stream gambar (JPG/PNG).
// Dilengkapi proteksi keamanan berlapis:
//  1. Pembatasan ukuran stream body maksimal 5 MB (Anti Memory Bomb).
//  2. Verifikasi DecodeConfig sebelum image.Decode() untuk menolak gambar > 4096 x 4096 px (Anti Decompression Bomb).
func DecodeImage(r io.Reader) (string, error) {
	// 1. Batasi pembacaan maksimal 5 MB + 1 byte untuk mendeteksi kelebihan ukuran
	limitedReader := io.LimitReader(r, MaxUploadSizeBytes+1)
	buf, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", err
	}

	if int64(len(buf)) > MaxUploadSizeBytes {
		return "", ErrImageTooLarge
	}

	// 2. Proteksi Decompression Bomb: Cek dimensi metadata terlebih dahulu
	configReader := bytes.NewReader(buf)
	cfg, _, err := image.DecodeConfig(configReader)
	if err != nil {
		return "", ErrInvalidPayload
	}

	if cfg.Width > MaxImageDimension || cfg.Height > MaxImageDimension {
		return "", ErrInvalidImageDimension
	}

	// 3. Decode seluruh gambar ke memori secara aman
	imageReader := bytes.NewReader(buf)
	img, _, err := image.Decode(imageReader)
	if err != nil {
		return "", err
	}

	// 4. Deteksi dan baca payload QR Code via gozxing
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return "", err
	}

	qrReader := gozxingqr.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return "", ErrNoQRCodeFound
	}

	return result.GetText(), nil
}
