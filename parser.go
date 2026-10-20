package bsdqris

import (
	"strconv"
	"strings"
)

// Parse mem-parse string QRIS EMVCo mentah menjadi objek QRISData.
// Sesuai Aturan 2: Integritas checksum CRC-16 divalidasi terlebih dahulu
// sebelum mengekstrak tag lainnya.
func Parse(rawPayload string) (*QRISData, error) {
	payload := strings.TrimSpace(rawPayload)
	if len(payload) < 30 {
		return nil, ErrPayloadTooShort
	}

	// 1. Gerbang Validasi Checksum CRC-16 (Strict Checksum Gate)
	if !VerifyCRC16(payload) {
		return nil, ErrInvalidCRC16
	}

	// 2. Token Traversal Loop untuk Root TLV Tags
	nodes := make([]TLVNode, 0)
	tags := make(map[string]string)
	idx := 0

	for idx < len(payload) {
		if idx+4 > len(payload) {
			return nil, ErrInvalidPayload
		}

		tag := payload[idx : idx+2]
		lenStr := payload[idx+2 : idx+4]
		length, err := strconv.Atoi(lenStr)
		if err != nil || length < 0 || idx+4+length > len(payload) {
			return nil, ErrInvalidPayload
		}

		val := payload[idx+4 : idx+4+length]
		nodes = append(nodes, TLVNode{
			Tag:    tag,
			Length: length,
			Value:  val,
		})
		tags[tag] = val

		idx += 4 + length
	}

	// 3. Validasi Versi Format (Tag 00 harus bernilai "01")
	if tags["00"] != "01" {
		return nil, ErrInvalidFormatVersion
	}

	// 4. Konstruksi Objek QRISData
	data := &QRISData{
		RawPayload:   payload,
		IsDynamic:    tags["01"] == "12",
		MerchantName: tags["59"],
		City:         tags["60"],
		PostalCode:   tags["61"],
		CurrencyCode: tags["53"],
		CRC16:        tags["63"],
		Tags:         tags,
		Nodes:        nodes,
	}

	// 5. Parse Amount (Tag 54) jika tersedia (QRIS Dinamis)
	if amtStr, ok := tags["54"]; ok && len(amtStr) > 0 {
		// Hilangkan desimal jika ada (contoh "45000.00" -> 45000)
		cleanAmt := strings.Split(amtStr, ".")[0]
		if amt, err := strconv.ParseInt(cleanAmt, 10, 64); err == nil {
			data.Amount = amt
		}
	}

	// 6. Deteksi Acquirer & NMID dari Tag 26 - 51
	data.AcquirerName, data.NMID = detectAcquirerAndNMID(tags)

	// 7. Ekstrak Terminal ID dari Tag 62 (Additional Data)
	if tag62, ok := tags["62"]; ok {
		subTags := ParseSubTLV(tag62)
		if tid, ok := subTags["07"]; ok && len(tid) > 0 {
			data.TerminalID = tid
		} else if tid, ok := subTags["01"]; ok && len(tid) > 0 {
			data.TerminalID = tid
		}
	}

	return data, nil
}

// ParseSubTLV mem-parse payload sub-tag TLV (seperti yang terdapat di dalam Tag 26-51 dan Tag 62)
func ParseSubTLV(subPayload string) map[string]string {
	result := make(map[string]string)
	idx := 0

	for idx+4 <= len(subPayload) {
		subTag := subPayload[idx : idx+2]
		lenStr := subPayload[idx+2 : idx+4]
		length, err := strconv.Atoi(lenStr)
		if err != nil || length < 0 || idx+4+length > len(subPayload) {
			break
		}

		subVal := subPayload[idx+4 : idx+4+length]
		result[subTag] = subVal
		idx += 4 + length
	}

	return result
}

// detectAcquirerAndNMID mendeteksi nama institusi penerbit/acquirer dan National Merchant ID (NMID)
func detectAcquirerAndNMID(tags map[string]string) (string, string) {
	var acquirerName string
	var nmid string

	// Prioritas 1: Periksa Tag 26 s/d Tag 45 (Merchant Account Information)
	for t := 26; t <= 45; t++ {
		tagKey := strconv.Itoa(t)
		val, exists := tags[tagKey]
		if !exists || len(val) < 4 {
			continue
		}

		subTags := ParseSubTLV(val)
		domain := strings.ToUpper(subTags["00"])

		// Cari NMID jika ada sub-tag berawalan ID...
		for _, subVal := range subTags {
			if strings.HasPrefix(strings.ToUpper(subVal), "ID1") || strings.HasPrefix(strings.ToUpper(subVal), "ID2") {
				nmid = subVal
			}
		}

		if strings.Contains(domain, "GOPAY") || strings.Contains(domain, "GOJEK") {
			acquirerName = "GoPay"
			break
		} else if strings.Contains(domain, "BCA") {
			acquirerName = "BCA"
			break
		} else if strings.Contains(domain, "MANDIRI") {
			acquirerName = "Bank Mandiri"
			break
		} else if strings.Contains(domain, "BRI") {
			acquirerName = "BRI"
			break
		} else if strings.Contains(domain, "BNI") {
			acquirerName = "BNI"
			break
		} else if strings.Contains(domain, "BSI") {
			acquirerName = "BSI"
			break
		} else if strings.Contains(domain, "SHOPEE") || strings.Contains(domain, "AIRPAY") {
			acquirerName = "ShopeePay"
			break
		} else if strings.Contains(domain, "OVO") {
			acquirerName = "OVO"
			break
		} else if strings.Contains(domain, "DANA") {
			acquirerName = "DANA"
			break
		} else if strings.Contains(domain, "NOBU") {
			acquirerName = "Bank Nobu"
			break
		} else if strings.Contains(domain, "NETZME") {
			acquirerName = "Netzme"
			break
		} else if acquirerName == "" && len(domain) > 0 {
			acquirerName = domain
		}
	}

	// Prioritas 2: Periksa Tag 51 (GPN National Identifier)
	if tag51, exists := tags["51"]; exists {
		subTags := ParseSubTLV(tag51)
		if nmid == "" {
			if idVal, ok := subTags["02"]; ok {
				nmid = idVal
			}
		}
		if acquirerName == "" {
			acquirerName = "GPN"
		}
	}

	if acquirerName == "" {
		acquirerName = "QRIS Standar Nasional"
	}

	return acquirerName, nmid
}
