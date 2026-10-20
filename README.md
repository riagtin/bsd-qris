# bsd-qris — The BSD QRIS Engine 🇮🇩

[![Go Report Card](https://goreportcard.com/badge/github.com/riagtin/bsd-qris)](https://goreportcard.com/report/github.com/riagtin/bsd-qris)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/riagtin/bsd-qris?status.svg)](https://pkg.go.dev/github.com/riagtin/bsd-qris)

Modul bersama resmi **Bintang Studio Digital** untuk pemrosesan, validasi, dan konversi **QRIS Standar Nasional (ASPI / Bank Indonesia)** berbasis spesifikasi **EMVCo Merchant-Presented Mode (MPM)**.

Modul ini dirancang dengan prinsip:
* **Zero Escrow / Zero Fee:** Menggunakan QRIS toko yang sudah ada, dana 100% langsung masuk ke rekening bank/e-wallet merchant tanpa potongan platform.
* **Token/Node Traversal:** Bebas dari bug kolisi substring (`strings.Index`) pada NMID dan MCC.
* **Strict Tag-62 Preservation:** Tag 62 bawaan acquirer (Terminal ID, Store Label) dijamin 100% utuh tanpa modifikasi.
* **Bank Indonesia Regulatory Limit:** Validasi otomatis batas nominal maksimal Rp 10.000.000 per transaksi (PADG No. 21/18/PADG/2019).
* **Anti Decompression Bomb:** Proteksi upload gambar hingga batas maksimal 4096 x 4096 piksel dan ukuran 5 MB.
* **Ultra-Fast Performance:** Eksekusi konversi $< 0.015\text{ ms}$ (lebih dari 90.000 operasi per detik).

---

## 📦 Instalasi

```bash
go get github.com/riagtin/bsd-qris
```

---

## 🚀 Penggunaan Cepat

### 1. Konversi QRIS Statis Menjadi QRIS Dinamis (Nominal Terkunci)

```go
package main

import (
	"fmt"
	"log"

	bsdqris "github.com/riagtin/bsd-qris"
)

func main() {
	staticQRIS := "00020101021126590016ID.CO.GOPAY.WWW...6304ABCD"
	amount := int64(47500) // Rp 47.500

	dynamicQRIS, err := bsdqris.GenerateDynamic(staticQRIS, amount)
	if err != nil {
		log.Fatalf("Gagal generate QRIS dinamis: %v", err)
	}

	fmt.Println("Payload QRIS Dinamis:", dynamicQRIS)
}
```

### 2. Parse & Validasi QRIS Toko

```go
data, err := bsdqris.Parse(rawPayload)
if err != nil {
    log.Fatalf("QRIS tidak valid: %v", err)
}

fmt.Printf("Merchant : %s\n", data.MerchantName)
fmt.Printf("Kota     : %s\n", data.City)
fmt.Printf("Acquirer : %s\n", data.AcquirerName)
fmt.Printf("NMID     : %s\n", data.NMID)
fmt.Printf("Terminal : %s\n", data.TerminalID)
```

### 3. Generate Gambar PNG untuk Nota PDF / WhatsApp

```go
pngBytes, err := bsdqris.ToPNG(dynamicQRIS, 256)
if err != nil {
    log.Fatal(err)
}
// pngBytes siap dilampirkan ke PDF Maroto v2 atau bot WhatsApp whatsmeow
```

### 4. Ekstraksi String dari Upload Foto Stiker QRIS

```go
payloadText, err := bsdqris.DecodeImage(fileReader)
if err != nil {
    log.Fatal(err)
}
```

---

## 🛡️ Benchmark Kinerja

Diuji pada prosesor AMD Ryzen 5:
```
BenchmarkParse-12              	  219782	      5097 ns/op (0.005 ms)
BenchmarkGenerateDynamic-12    	  110308	     11125 ns/op (0.011 ms)
```

---

## 📄 Lisensi
Hak Cipta © 2026 Bintang Studio Digital.  
Didistribusikan di bawah Lisensi MIT. Lihat file [LICENSE](LICENSE) untuk informasi lebih lanjut.
