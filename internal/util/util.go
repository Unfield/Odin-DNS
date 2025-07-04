package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Unfield/Odin-DNS/pkg/odintypes"
)

func ParseDomainName(buffer []byte, offset int) (string, int, error) {
	var name string
	originalOffset := offset

	for {
		if offset >= len(buffer) {
			return "", originalOffset, fmt.Errorf("buffer too short for domain name")
		}

		length := buffer[offset]
		offset++

		if length == 0 {
			break
		}

		if (length & 0xC0) == 0xC0 {
			if offset >= len(buffer) {
				return "", originalOffset, fmt.Errorf("buffer too short for domain name pointer")
			}
			pointerOffset := (int(length&0x3F) << 8) | int(buffer[offset])
			offset++

			if pointerOffset >= len(buffer) {
				return "", originalOffset, fmt.Errorf("pointer out of bounds in domain name")
			}

			pointedToName, _, err := ParseDomainName(buffer, pointerOffset)
			if err != nil {
				return "", originalOffset, fmt.Errorf("failed to resolve pointer: %w", err)
			}
			name += pointedToName

			if len(name) > 0 && name[len(name)-1] == '.' {
				name = name[:len(name)-1]
			}
			return name, offset, nil
		}

		if offset+int(length) > len(buffer) {
			return "", originalOffset, fmt.Errorf("buffer too short for domain label")
		}

		label := buffer[offset : offset+int(length)]
		name += string(label) + "."
		offset += int(length)
	}

	if len(name) > 0 && name[len(name)-1] == '.' {
		name = name[:len(name)-1]
	}

	return name, offset, nil
}

func ParseType(typeCode uint16) (string, error) {
	switch typeCode {
	case 1:
		return "A", nil
	case 2:
		return "NS", nil
	case 5:
		return "CNAME", nil
	case 6:
		return "SOA", nil
	case 12:
		return "PTR", nil
	case 15:
		return "MX", nil
	case 16:
		return "TXT", nil
	case 28:
		return "AAAA", nil
	default:
		return "", fmt.Errorf("unknown type code: %d", typeCode)
	}
}

func ParseTypeOrNA(typeCode uint16) string {
	pt, err := ParseType(typeCode)
	if err != nil {
		return "N/A"
	}
	return pt
}

func ParseClass(classCode uint16) (string, error) {
	switch classCode {
	case 1:
		return "IN", nil
	default:
		return "", fmt.Errorf("unknown class code: %d", classCode)
	}
}

func ParseFlags(flags uint16) odintypes.DNSHeaderFlags {
	return odintypes.DNSHeaderFlags{
		QR:     (flags & 0x8000) != 0,
		Opcode: uint8((flags & 0x7800) >> 11),
		AA:     (flags & 0x0400) != 0,
		TC:     (flags & 0x0200) != 0,
		RD:     (flags & 0x0100) != 0,
		RA:     (flags & 0x0080) != 0,
		Z:      uint8((flags & 0x0070) >> 4),
		RCode:  uint8(flags & 0x000F),
	}
}

func FormatDomainName(name string) []byte {
	if name == "" {
		return []byte{0}
	}

	labels := splitDomainName(name)
	var result []byte

	for _, label := range labels {
		if len(label) > 63 {
			continue
		}
		result = append(result, byte(len(label)))
		result = append(result, label...)
	}

	result = append(result, 0)
	return result
}

func splitDomainName(name string) []string {
	labels := []string{}
	start := 0

	for i, char := range name {
		if char == '.' {
			if start < i {
				labels = append(labels, name[start:i])
			}
			start = i + 1
		}
	}

	if start < len(name) {
		labels = append(labels, name[start:])
	}

	return labels
}

type CheckForDemoFailedResponse struct {
	Message string `json:"message"`
}

func CheckForDemoKey(queryParams url.Values, w http.ResponseWriter, demoKey string) bool {
	if queryParams.Get("demo_key") == "" || queryParams.Get("demo_key") != demoKey {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		response := CheckForDemoFailedResponse{
			Message: "Invalid or missing demo key",
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return false
		}
		return false
	}
	return true
}

func ConvertRDataStringToBytes(recordType uint16, rDataString string) ([]byte, error) {
	switch recordType {
	case odintypes.TYPE_A:
		return odintypes.ParseA_RData(rDataString)
	case odintypes.TYPE_AAAA:
		return odintypes.ParseAAAA_RData(rDataString)
	case odintypes.TYPE_CNAME, odintypes.TYPE_NS, odintypes.TYPE_PTR:
		return odintypes.ParseDomainName_RData(rDataString)
	case odintypes.TYPE_MX:
		return odintypes.ParseMX_RData(rDataString)
	case odintypes.TYPE_TXT:
		return odintypes.ParseTXT_RData(rDataString)
	default:
		return nil, fmt.Errorf("unsupported RData conversion for record type %d", recordType)
	}
}

func ConvertRDataBytesToString(recordType uint16, rDataBytes []byte) string {
	switch recordType {
	case odintypes.TYPE_A:
		return odintypes.FormatA_RData(rDataBytes)
	case odintypes.TYPE_AAAA:
		return odintypes.FormatAAAA_RData(rDataBytes)
	case odintypes.TYPE_CNAME, odintypes.TYPE_NS, odintypes.TYPE_PTR:
		return odintypes.FormatDomainName_RData(rDataBytes)
	case odintypes.TYPE_MX:
		return odintypes.FormatMX_RData(rDataBytes)
	case odintypes.TYPE_TXT:
		return odintypes.FormatTXT_RData(rDataBytes)
	default:
		return fmt.Sprintf("Unsupported_RData_Format_%d", recordType)
	}
}

func RespondWithJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
