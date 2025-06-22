package parser

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/Unfield/Odin-DNS/pkg/odintypes"
)

func PackResponse(response *odintypes.DNSRequest) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, response.Header.ID); err != nil {
		return nil, fmt.Errorf("failed to pack header ID: %w", err)
	}

	flags := response.Header.Flags.ToUint16()
	if err := binary.Write(buf, binary.BigEndian, flags); err != nil {
		return nil, fmt.Errorf("failed to pack header flags: %w", err)
	}

	if err := binary.Write(buf, binary.BigEndian, response.Header.QDCount); err != nil {
		return nil, fmt.Errorf("failed to pack header QDCount: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, response.Header.ANCount); err != nil {
		return nil, fmt.Errorf("failed to pack header ANCount: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, response.Header.NSCount); err != nil {
		return nil, fmt.Errorf("failed to pack header NSCount: %w", err)
	}
	if err := binary.Write(buf, binary.BigEndian, response.Header.ARCount); err != nil {
		return nil, fmt.Errorf("failed to pack header ARCount: %w", err)
	}

	nameOffsets := make(map[string]uint16)
	currentOffset := uint16(12)

	for _, q := range response.Questions {
		packedName, err := packDomainName(q.Name, nameOffsets, buf.Len())
		if err != nil {
			return nil, fmt.Errorf("failed to pack question name '%s': %w", q.Name, err)
		}
		if _, err := buf.Write(packedName); err != nil {
			return nil, fmt.Errorf("failed to write packed question name: %w", err)
		}

		if err := binary.Write(buf, binary.BigEndian, q.Type); err != nil {
			return nil, fmt.Errorf("failed to pack question type: %w", err)
		}
		if err := binary.Write(buf, binary.BigEndian, q.Class); err != nil {
			return nil, fmt.Errorf("failed to pack question class: %w", err)
		}
		currentOffset += uint16(len(packedName) + 4)
	}

	for _, a := range response.Answers {
		packedName, err := packDomainName(a.Name, nameOffsets, buf.Len())
		if err != nil {
			return nil, fmt.Errorf("failed to pack answer name '%s': %w", a.Name, err)
		}
		if _, err := buf.Write(packedName); err != nil {
			return nil, fmt.Errorf("failed to write packed answer name: %w", err)
		}

		if err := binary.Write(buf, binary.BigEndian, a.Type); err != nil {
			return nil, fmt.Errorf("failed to pack answer type: %w", err)
		}
		if err := binary.Write(buf, binary.BigEndian, a.Class); err != nil {
			return nil, fmt.Errorf("failed to pack answer class: %w", err)
		}
		if err := binary.Write(buf, binary.BigEndian, a.TTL); err != nil {
			return nil, fmt.Errorf("failed to pack answer TTL: %w", err)
		}

		rdLengthPos := buf.Len()
		if err := binary.Write(buf, binary.BigEndian, uint16(0)); err != nil {
			return nil, fmt.Errorf("failed to write RDLENGTH placeholder: %w", err)
		}

		rdataStartPos := buf.Len()

		if err := packRData(a.Type, a.RData, buf, nameOffsets); err != nil {
			return nil, fmt.Errorf("failed to pack RData for type %d: %w", a.Type, err)
		}

		rdataLen := uint16(buf.Len() - rdataStartPos)
		binary.BigEndian.PutUint16(buf.Bytes()[rdLengthPos:], rdataLen)
	}

	return buf.Bytes(), nil
}

func packDomainName(domain string, nameOffsets map[string]uint16, currentBufferLen int) ([]byte, error) {
	if offset, ok := nameOffsets[domain]; ok {
		pointer := uint16(0xC000) | offset
		buf := new(bytes.Buffer)
		if err := binary.Write(buf, binary.BigEndian, pointer); err != nil {
			return nil, fmt.Errorf("failed to write domain name pointer: %w", err)
		}
		return buf.Bytes(), nil
	}

	var packedName []byte
	parts := strings.Split(domain, ".")

	if domain != "" && domain != "." {
		nameOffsets[domain] = uint16(currentBufferLen)
	}

	for _, part := range parts {
		if part == "" {
			if len(parts) == 1 && domain == "." {
				packedName = append(packedName, 0x00)
				break
			}
			continue
		}
		if len(part) > 63 {
			return nil, fmt.Errorf("DNS label '%s' too long (max 63 characters)", part)
		}
		packedName = append(packedName, byte(len(part)))
		packedName = append(packedName, []byte(part)...)

		suffix := strings.Join(parts[1:], ".")
		if suffix != "" {
		}
	}
	packedName = append(packedName, 0x00)

	return packedName, nil
}

func packRData(recordType uint16, rData []byte, buf *bytes.Buffer, nameOffsets map[string]uint16) error {
	switch recordType {
	case odintypes.TYPE_A:
		if len(rData) != 4 {
			return fmt.Errorf("A record RData must be 4 bytes for IPv4, got %d", len(rData))
		}
		if _, err := buf.Write(rData); err != nil {
			return fmt.Errorf("failed to write A record RData: %w", err)
		}
	case odintypes.TYPE_AAAA:
		if len(rData) != 16 {
			return fmt.Errorf("AAAA record RData must be 16 bytes for IPv6, got %d", len(rData))
		}
		if _, err := buf.Write(rData); err != nil {
			return fmt.Errorf("failed to write AAAA record RData: %w", err)
		}
	case odintypes.TYPE_CNAME, odintypes.TYPE_NS, odintypes.TYPE_PTR:
		domainNameString := string(rData)
		packedDomain, err := packDomainName(domainNameString, nameOffsets, buf.Len())
		if err != nil {
			return fmt.Errorf("failed to pack RData domain name '%s': %w", domainNameString, err)
		}
		if _, err := buf.Write(packedDomain); err != nil {
			return fmt.Errorf("failed to write packed RData domain name: %w", err)
		}

	case odintypes.TYPE_MX:
		if len(rData) < 2 {
			return fmt.Errorf("MX record RData too short, must contain preference: got %d bytes", len(rData))
		}
		preference := binary.BigEndian.Uint16(rData[0:2])
		if err := binary.Write(buf, binary.BigEndian, preference); err != nil {
			return fmt.Errorf("failed to write MX preference: %w", err)
		}

		domainNameString := string(rData[2:])
		if domainNameString == "" {
			return fmt.Errorf("MX record RData has no domain name after preference")
		}

		packedDomain, err := packDomainName(domainNameString, nameOffsets, buf.Len())
		if err != nil {
			return fmt.Errorf("failed to pack MX RData domain name '%s': %w", domainNameString, err)
		}
		if _, err := buf.Write(packedDomain); err != nil {
			return fmt.Errorf("failed to write packed MX RData domain name: %w", err)
		}

	case odintypes.TYPE_TXT:
		textBytes := rData
		if len(textBytes) > 255 {
			return fmt.Errorf("TXT record string too long (max 255 bytes per string segment): %d", len(textBytes))
		}
		if err := binary.Write(buf, binary.BigEndian, byte(len(textBytes))); err != nil {
			return fmt.Errorf("failed to write TXT RData length: %w", err)
		}
		if _, err := buf.Write(textBytes); err != nil {
			return fmt.Errorf("failed to write TXT RData: %w", err)
		}

	default:
		if _, err := buf.Write(rData); err != nil {
			return fmt.Errorf("failed to write generic RData for type %d: %w", recordType, err)
		}
	}
	return nil
}
