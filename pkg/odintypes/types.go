package odintypes

import (
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
)

type DNSRequest struct {
	Header     DNSHeader
	Questions  []DNSQuestion
	Answers    []*DNSRecord
	Authority  []*DNSRecord
	Additional []*DNSRecord
}

type DNSHeader struct {
	ID      uint16
	Flags   DNSHeaderFlags
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

type DNSHeaderFlags struct {
	QR     bool
	Opcode uint8
	AA     bool
	TC     bool
	RD     bool
	RA     bool
	Z      uint8
	AD     bool
	CD     bool
	RCode  uint8
}

func (f *DNSHeaderFlags) ToUint16() uint16 {
	var flags uint16
	if f.QR {
		flags |= (1 << 15)
	}
	flags |= (uint16(f.Opcode) & 0xF) << 11
	if f.AA {
		flags |= (1 << 10)
	}
	if f.TC {
		flags |= (1 << 9)
	}
	if f.RD {
		flags |= (1 << 8)
	}
	if f.RA {
		flags |= (1 << 7)
	}
	flags |= (uint16(f.Z) & 0x7) << 4

	if f.AD {
		flags |= (1 << 5)
	}
	if f.CD {
		flags |= (1 << 4)
	}

	flags |= (uint16(f.RCode) & 0xF)

	return flags
}

type DNSQuestion struct {
	Name  string
	Type  uint16
	Class uint16
}

type DNSRecord struct {
	Name  string
	Type  uint16
	Class uint16
	TTL   uint32
	RData []byte
}

const (
	TYPE_A     uint16 = 1
	TYPE_NS    uint16 = 2
	TYPE_CNAME uint16 = 5
	TYPE_SOA   uint16 = 6
	TYPE_MX    uint16 = 15
	TYPE_TXT   uint16 = 16
	TYPE_AAAA  uint16 = 28
	TYPE_SRV   uint16 = 33
	TYPE_PTR   uint16 = 12
	TYPE_ANY   uint16 = 255

	CLASS_IN uint16 = 1
)

func StringToType(s string) (uint16, error) {
	switch s {
	case "A":
		return TYPE_A, nil
	case "NS":
		return TYPE_NS, nil
	case "CNAME":
		return TYPE_CNAME, nil
	case "SOA":
		return TYPE_SOA, nil
	case "MX":
		return TYPE_MX, nil
	case "TXT":
		return TYPE_TXT, nil
	case "AAAA":
		return TYPE_AAAA, nil
	case "SRV":
		return TYPE_SRV, nil
	case "PTR":
		return TYPE_PTR, nil
	case "ANY":
		return TYPE_ANY, nil
	default:
		if i, err := strconv.ParseUint(s, 10, 16); err == nil {
			return uint16(i), nil
		}
		return 0, fmt.Errorf("unknown DNS record type: %s", s)
	}
}

func TypeToString(t uint16) string {
	switch t {
	case TYPE_A:
		return "A"
	case TYPE_NS:
		return "NS"
	case TYPE_CNAME:
		return "CNAME"
	case TYPE_SOA:
		return "SOA"
	case TYPE_MX:
		return "MX"
	case TYPE_TXT:
		return "TXT"
	case TYPE_AAAA:
		return "AAAA"
	case TYPE_SRV:
		return "SRV"
	case TYPE_PTR:
		return "PTR"
	case TYPE_ANY:
		return "ANY"
	default:
		return fmt.Sprintf("TYPE%d", t)
	}
}

func StringToClass(s string) (uint16, error) {
	switch s {
	case "IN":
		return CLASS_IN, nil
	default:
		if i, err := strconv.ParseUint(s, 10, 16); err == nil {
			return uint16(i), nil
		}
		return 0, fmt.Errorf("unknown DNS record class: %s", s)
	}
}

func ClassToString(c uint16) string {
	switch c {
	case CLASS_IN:
		return "IN"
	default:
		return fmt.Sprintf("CLASS%d", c)
	}
}

func ParseA_RData(s string) ([]byte, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("invalid IPv4 address format: %s", s)
	}
	ipv4 := ip.To4()
	if ipv4 == nil {
		return nil, fmt.Errorf("not an IPv4 address: %s", s)
	}
	return ipv4, nil
}

func ParseAAAA_RData(s string) ([]byte, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, fmt.Errorf("invalid IPv6 address string: %s", s)
	}
	ipv6 := ip.To16()
	if ipv6 == nil || ip.To4() != nil && len(s) < 16 {
		return nil, fmt.Errorf("address '%s' is not a valid IPv6 address for AAAA record", s)
	}
	return ipv6, nil
}

func ParseDomainName_RData(s string) ([]byte, error) {
	if len(s) == 0 {
		return nil, fmt.Errorf("domain name cannot be empty")
	}
	if strings.ContainsAny(s, " \t\n\r") {
		return nil, fmt.Errorf("domain name contains invalid characters: %s", s)
	}
	return []byte(s), nil
}

func ParseMX_RData(s string) ([]byte, error) {
	parts := strings.SplitN(s, " ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid MX record RData format, expected 'PREFERENCE DOMAIN.NAME': %s", s)
	}

	pref, err := strconv.ParseUint(parts[0], 10, 16)
	if err != nil {
		return nil, fmt.Errorf("invalid MX preference '%s': %w", parts[0], err)
	}

	domainName := parts[1]
	if len(domainName) == 0 {
		return nil, fmt.Errorf("MX domain name cannot be empty")
	}

	prefBytes := make([]byte, 2)
	binary.BigEndian.PutUint16(prefBytes, uint16(pref))

	return append(prefBytes, []byte(domainName)...), nil
}

func ParseTXT_RData(s string) ([]byte, error) {
	if len(s) > 255 {
		return nil, fmt.Errorf("TXT record string is too long (max 255 bytes): %d bytes", len(s))
	}
	return []byte(s), nil
}
