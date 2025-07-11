package parser

import (
	"fmt"

	"github.com/Unfield/Odin-DNS/internal/util"
	"github.com/Unfield/Odin-DNS/pkg/odintypes"
)

func ParseRequest(buffer []byte) (odintypes.DNSRequest, error) {
	if len(buffer) < 12 {
		return odintypes.DNSRequest{}, fmt.Errorf("buffer too short to contain DNS header")
	}

	var request odintypes.DNSRequest
	headerSection := [12]byte{}
	copy(headerSection[:], buffer[:12])
	header, err := ParseHeaderSection(headerSection)
	if err != nil {
		return odintypes.DNSRequest{}, fmt.Errorf("error parsing header section: %w", err)
	}
	request.Header = header

	offset := 12
	var qsection []odintypes.DNSQuestion
	for i := range int(header.QDCount) {
		if offset >= len(buffer) {
			return odintypes.DNSRequest{}, fmt.Errorf("buffer too short for question section %d", i+1)
		}
		q, newOffset, err := ParseQuestionSection(buffer, offset)
		if err != nil {
			return odintypes.DNSRequest{}, fmt.Errorf("error parsing question section %d: %w", i+1, err)
		}
		offset = newOffset

		qsection = append(qsection, q)
	}

	request.Questions = qsection

	return request, nil
}

func ParseHeaderSection(headerSection [12]byte) (odintypes.DNSHeader, error) {
	var hsection odintypes.DNSHeader

	hsection.ID = uint16(headerSection[0])<<8 | uint16(headerSection[1])
	hsection.Flags = util.ParseFlags(uint16(headerSection[2])<<8 | uint16(headerSection[3]))
	hsection.QDCount = uint16(headerSection[4])<<8 | uint16(headerSection[5])
	hsection.ANCount = uint16(headerSection[6])<<8 | uint16(headerSection[7])
	hsection.NSCount = uint16(headerSection[8])<<8 | uint16(headerSection[9])
	hsection.ARCount = uint16(headerSection[10])<<8 | uint16(headerSection[11])

	return hsection, nil
}

func ParseQuestionSection(buffer []byte, offset int) (odintypes.DNSQuestion, int, error) {
	if offset+4 > len(buffer) {
		return odintypes.DNSQuestion{}, offset, fmt.Errorf("buffer too short for question section")
	}

	var qsection odintypes.DNSQuestion
	name, newOffset, err := util.ParseDomainName(buffer, offset)
	if err != nil {
		return odintypes.DNSQuestion{}, newOffset, fmt.Errorf("error parsing domain name: %w", err)
	}
	qsection.Name = name

	if newOffset+4 > len(buffer) {
		return odintypes.DNSQuestion{}, newOffset, fmt.Errorf("buffer too short for question type and class")
	}
	qsection.Type = uint16(buffer[newOffset])<<8 | uint16(buffer[newOffset+1])
	newOffset += 2

	qsection.Class = uint16(buffer[newOffset])<<8 | uint16(buffer[newOffset+1])
	newOffset += 2

	return qsection, newOffset, nil
}
