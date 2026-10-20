package bsdqris

import "errors"

// Konstanta batas regulasi Bank Indonesia & keamanan sistem
const (
	// MinQRISAmount adalah nilai transaksi terkecil yang valid (Rp 1)
	MinQRISAmount int64 = 1

	// MaxQRISAmount adalah batas maksimum 1 kali transaksi QRIS sesuai regulasi Bank Indonesia (PADG No. 21/18/PADG/2019)
	MaxQRISAmount int64 = 10000000 // Rp 10.000.000 (Sepuluh Juta Rupiah)

	// MaxUploadSizeBytes adalah batas maksimal ukuran file upload foto stiker QRIS (5 MB)
	MaxUploadSizeBytes int64 = 5 << 20 // 5 MB

	// MaxImageDimension adalah batas resolusi piksel lebar/tinggi gambar untuk mencegah Decompression Bomb
	MaxImageDimension int = 4096
)

// Definisi error domain bsd-qris
var (
	ErrInvalidCRC16          = errors.New("bsd-qris: checksum CRC16 tidak valid atau payload terpotong")
	ErrInvalidPayload        = errors.New("bsd-qris: format string QRIS tidak valid atau bukan format EMVCo")
	ErrPayloadTooShort       = errors.New("bsd-qris: panjang payload QRIS terlalu pendek")
	ErrInvalidFormatVersion  = errors.New("bsd-qris: versi format QRIS (Tag 00) bukan '01'")
	ErrInvalidAmount         = errors.New("bsd-qris: nominal harus lebih besar dari 0 (minimal Rp 1)")
	ErrExceedsQRISLimit      = errors.New("bsd-qris: nominal melampaui batas transaksi maksimal QRIS Bank Indonesia (Rp 10.000.000)")
	ErrImageTooLarge         = errors.New("bsd-qris: ukuran file gambar melebihi batas 5 MB")
	ErrInvalidImageDimension = errors.New("bsd-qris: dimensi gambar melebihi batas aman 4096 x 4096 piksel")
	ErrNoQRCodeFound         = errors.New("bsd-qris: tidak ditemukan QR Code yang dapat dibaca pada gambar yang diunggah")
	ErrTagNotFound           = errors.New("bsd-qris: tag yang dicari tidak ditemukan dalam payload")
)

// TLVNode merepresentasikan satu elemen Tag-Length-Value pada QRIS EMVCo
type TLVNode struct {
	Tag    string `json:"tag"`
	Length int    `json:"length"`
	Value  string `json:"value"`
}

// QRISData merepresentasikan hasil dekonstruksi data QRIS lengkap
type QRISData struct {
	RawPayload   string            `json:"raw_payload"`
	IsDynamic    bool              `json:"is_dynamic"`
	AcquirerName string            `json:"acquirer_name"`
	NMID         string            `json:"nmid,omitempty"`
	MerchantName string            `json:"merchant_name"`
	City         string            `json:"city"`
	PostalCode   string            `json:"postal_code"`
	CurrencyCode string            `json:"currency_code"` // "360" = IDR
	Amount       int64             `json:"amount,omitempty"`
	TerminalID   string            `json:"terminal_id,omitempty"`
	CRC16        string            `json:"crc16"`
	Tags         map[string]string `json:"tags"`
	Nodes        []TLVNode         `json:"-"`
}
