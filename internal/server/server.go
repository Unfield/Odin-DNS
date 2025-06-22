package server

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/Unfield/Odin-DNS/internal/config"
	mysql "github.com/Unfield/Odin-DNS/internal/datastore/MySQL"
	"github.com/Unfield/Odin-DNS/internal/parser"
	"github.com/Unfield/Odin-DNS/pkg/odintypes"
)

func StartServer(config *config.Config) {
	logger := slog.Default().WithGroup("DNS-Server")

	mysqlDriver, err := mysql.NewMySQLDriver(config.MySQL_DSN)
	if err != nil {
		logger.Error("Failed to connect to MySQL", "error", err)
		return
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", config.DNS_PORT))
	if err != nil {
		logger.Error("Error resolving address", "port", config.DNS_PORT, "error", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Error("Error listening on UDP port", "port", config.DNS_PORT, "error", err)
		return
	}
	defer conn.Close()

	logger.Info("Odin DNS server is running", "port", addr.Port)

	buffer := make([]byte, config.BUFFER_SIZE)

	for {
		_, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			logger.Error("Error reading from UDP", "error", err)
			continue
		}

		response := &odintypes.DNSRequest{
			Header: odintypes.DNSHeader{
				ID: 0,
				Flags: odintypes.DNSHeaderFlags{
					QR: true,
					RA: false,
					Z:  0,
				},
				QDCount: 0,
				ANCount: 0,
				NSCount: 0,
				ARCount: 0,
			},
			Questions:  []odintypes.DNSQuestion{},
			Answers:    []*odintypes.DNSRecord{},
			Authority:  []*odintypes.DNSRecord{},
			Additional: []*odintypes.DNSRecord{},
		}

		req, parseErr := parser.ParseRequest(buffer)
		if parseErr != nil {
			logger.Error("Error parsing DNS request", "error", parseErr, "client", clientAddr.String())
			response.Header.Flags.RCode = 1
			if sendErr := SendResponse(response, conn, clientAddr); sendErr != nil {
				logger.Error("Error sending FORMERR response", "error", sendErr)
			}
			continue
		}

		response.Header.ID = req.Header.ID
		response.Header.QDCount = req.Header.QDCount
		response.Questions = req.Questions
		response.Header.Flags.RCode = 0

		logger.Debug("Received DNS request", "client", clientAddr.String(), "request", req)

		if req.Header.Flags.QR {
			//TODO: Handle response
			logger.Warn("Received a response instead of a request")
			continue
		}

		if len(req.Questions) == 0 {
			logger.Warn("Received request with no questions", "client", clientAddr.String(), "id", req.Header.ID)
			response.Header.Flags.RCode = 1
			if sendErr := SendResponse(response, conn, clientAddr); sendErr != nil {
				logger.Error("Error sending NoQuestions response", "error", sendErr)
			}
			continue
		}

		logger.Info("Processing DNS request", "domain", req.Questions[0].Name)

		for _, question := range req.Questions {
			dnsRecord, err := mysqlDriver.LookupRecordForDNSQuery(question.Name, question.Type, question.Class)

			if err != nil {
				logger.Error("Database lookup error", "name", question.Name, "type", question.Type, "class", question.Class, "error", err)
				response.Header.Flags.RCode = 2
				response.Answers = []*odintypes.DNSRecord{}
				response.Authority = []*odintypes.DNSRecord{}
				response.Additional = []*odintypes.DNSRecord{}
				response.Header.ANCount = 0
				response.Header.NSCount = 0
				response.Header.ARCount = 0
				if sendErr := SendResponse(response, conn, clientAddr); sendErr != nil {
					logger.Error("Error sending SERVFAIL response", "error", sendErr)
				}
				goto EndCurrentRequest
			}

			if dnsRecord == nil {
				logger.Warn("Resource Record not found (from DB)", "name", question.Name, "type", question.Type, "class", question.Class, "client", clientAddr.String(), "id", req.Header.ID)
				response.Header.Flags.RCode = 3
				response.Answers = []*odintypes.DNSRecord{}
				response.Authority = []*odintypes.DNSRecord{}
				response.Additional = []*odintypes.DNSRecord{}
				response.Header.ANCount = 0
				response.Header.NSCount = 0
				response.Header.ARCount = 0
				if sendErr := SendResponse(response, conn, clientAddr); sendErr != nil {
					logger.Error("Error sending NXDOMAIN response", "error", sendErr)
				}
				goto EndCurrentRequest
			}

			response.Answers = append(response.Answers, dnsRecord)
			response.Header.ANCount = response.Header.ANCount + 1
			response.Header.Flags.AA = true
		}

		err = SendResponse(response, conn, clientAddr)
		if err != nil {
			logger.Error("Error packing NXDOMAIN response", "error", err)
		}

	EndCurrentRequest:
	}

}

func SendResponse(response *odintypes.DNSRequest, conn *net.UDPConn, clientAddr *net.UDPAddr) error {
	binaryResponse, err := parser.PackResponse(response)
	if err != nil {
		return fmt.Errorf("Error sending NXDOMAIN response: %s", err)
	}
	_, err = conn.WriteToUDP(binaryResponse, clientAddr)
	if err != nil {
		return fmt.Errorf("Error sending NXDOMAIN response: %s", err)
	}
	return nil
}
