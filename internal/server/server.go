package server

import (
	"fmt"
	"log/slog"
	"net"
	"time"

	"github.com/Unfield/Odin-DNS/internal/config"
	mysql "github.com/Unfield/Odin-DNS/internal/datastore/MySQL"
	redis "github.com/Unfield/Odin-DNS/internal/datastore/Redis"
	"github.com/Unfield/Odin-DNS/internal/metrics"
	"github.com/Unfield/Odin-DNS/internal/parser"
	"github.com/Unfield/Odin-DNS/internal/util"
	"github.com/Unfield/Odin-DNS/pkg/odintypes"
)

func StartServer(config *config.Config) {
	logger := slog.Default().WithGroup("DNS-Server")

	logger.Info("Initializing metrics ingestion driver...")
	ingestionDriver := metrics.NewClickHouseIngestionDriver(config)
	logger.Info("Metrics ingestion driver initialized and batch processing started.")
	defer func() {
		logger.Info("Closing ingestion driver...")
		if err := ingestionDriver.Close(); err != nil {
			logger.Error("Error closing ingestion driver", "error", err)
		} else {
			logger.Info("Ingestion driver closed successfully.")
		}
	}()

	mysqlDriver, err := mysql.NewMySQLDriver(config.MySQL_DSN)
	if err != nil {
		logger.Error("Failed to connect to MySQL", "error", err)
		return
	}

	cacheDriver := redis.NewRedisCacheDriver(mysqlDriver, config.REDIS_HOST, config.REDIS_USERNAME, config.REDIS_PASSWORD, config.REDIS_DATABASE)

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", config.DNS_HOST, config.DNS_PORT))
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

		startTime := time.Now()

		currentMetric := metrics.DNSMetric{
			Timestamp: time.Now(),
			IP:        clientAddr.IP.String(),
			Success:   1,
			Domain:    "N/A",
			QueryType: "N/A",
			CacheHit:  0,
			Rcode:     0,
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

			currentMetric.Success = 0
			currentMetric.ErrorMessage = fmt.Sprintf("FORMERR: %v", parseErr)
			currentMetric.Rcode = response.Header.Flags.RCode

			if sendErr := SendResponse(response, conn, clientAddr); sendErr != nil {
				logger.Error("Error sending FORMERR response", "error", sendErr)
			}
			currentMetric.ResponseTimeMs = float64(time.Since(startTime).Milliseconds())
			ingestionDriver.Collect(currentMetric)
			continue
		}

		response.Header.ID = req.Header.ID
		response.Header.QDCount = req.Header.QDCount
		response.Questions = req.Questions

		logger.Debug("Received DNS request", "client", clientAddr.String(), "request", req)

		if req.Header.Flags.QR {
			logger.Warn("Received a response instead of a query; ignoring.", "client", clientAddr.String(), "id", req.Header.ID)
			continue
		}

		if len(req.Questions) > 0 {
			currentMetric.Domain = req.Questions[0].Name
			currentMetric.QueryType = util.ParseTypeOrNA(req.Questions[0].Type)
		} else {
			logger.Warn("Received request with no questions", "client", clientAddr.String(), "id", req.Header.ID)
			response.Header.Flags.RCode = 1
			response.Header.Flags.QR = true

			currentMetric.Success = 0
			currentMetric.ErrorMessage = "FORMERR: No questions in request"
			currentMetric.Rcode = response.Header.Flags.RCode

			if sendErr := SendResponse(response, conn, clientAddr); sendErr != nil {
				logger.Error("Error sending NoQuestions response", "error", sendErr)
			}
			currentMetric.ResponseTimeMs = float64(time.Since(startTime).Milliseconds())
			ingestionDriver.Collect(currentMetric)
			continue
		}

		logger.Info("Processing DNS request", "domain", currentMetric.Domain, "type", currentMetric.QueryType)

		var cacheHit uint8

		question := req.Questions[0]

		dnsRecord, cacheHit, err := cacheDriver.LookupRecordForDNSQuery(question.Name, question.Type, question.Class)
		currentMetric.CacheHit = cacheHit

		if err != nil {
			logger.Error("Database lookup error", "name", question.Name, "type", question.Type, "class", question.Class, "error", err)
			response.Header.Flags.RCode = 2

			currentMetric.Success = 0
			currentMetric.ErrorMessage = fmt.Sprintf("SERVFAIL: %v", err)
			currentMetric.Rcode = response.Header.Flags.RCode

			response.Answers = []*odintypes.DNSRecord{}
			response.Authority = []*odintypes.DNSRecord{}
			response.Additional = []*odintypes.DNSRecord{}
			response.Header.ANCount = 0
			response.Header.NSCount = 0
			response.Header.ARCount = 0
			if sendErr := SendResponse(response, conn, clientAddr); sendErr != nil {
				logger.Error("Error sending SERVFAIL response", "error", sendErr)
			}
			currentMetric.ResponseTimeMs = float64(time.Since(startTime).Milliseconds())
			ingestionDriver.Collect(currentMetric)
			goto EndCurrentRequest
		}

		if dnsRecord == nil {
			logger.Warn("Resource Record not found (from DB)", "name", question.Name, "type", question.Type, "class", question.Class, "client", clientAddr.String(), "id", req.Header.ID)
			response.Header.Flags.RCode = 3

			currentMetric.Success = 0
			currentMetric.ErrorMessage = "NXDOMAIN: Record not found"
			currentMetric.Rcode = response.Header.Flags.RCode

			response.Answers = []*odintypes.DNSRecord{}
			response.Authority = []*odintypes.DNSRecord{}
			response.Additional = []*odintypes.DNSRecord{}
			response.Header.ANCount = 0
			response.Header.NSCount = 0
			response.Header.ARCount = 0
			if sendErr := SendResponse(response, conn, clientAddr); sendErr != nil {
				logger.Error("Error sending NXDOMAIN response", "error", sendErr)
			}
			currentMetric.ResponseTimeMs = float64(time.Since(startTime).Milliseconds())
			ingestionDriver.Collect(currentMetric)
			goto EndCurrentRequest
		}

		response.Answers = append(response.Answers, dnsRecord)
		response.Header.ANCount = response.Header.ANCount + 1
		response.Header.Flags.AA = true

		currentMetric.Success = 1
		currentMetric.ErrorMessage = ""
		currentMetric.Rcode = response.Header.Flags.RCode

		if sendErr := SendResponse(response, conn, clientAddr); sendErr != nil {
			logger.Error("Error sending DNS response", "error", sendErr, "client", clientAddr.String(), "domain", currentMetric.Domain)
			currentMetric.Success = 0
			currentMetric.ErrorMessage = fmt.Sprintf("SendResponse failed: %v", sendErr)
			currentMetric.Rcode = 2
		}

		currentMetric.ResponseTimeMs = float64(time.Since(startTime).Milliseconds())
		ingestionDriver.Collect(currentMetric)

	EndCurrentRequest:
	}

}

func SendResponse(response *odintypes.DNSRequest, conn *net.UDPConn, clientAddr *net.UDPAddr) error {
	binaryResponse, err := parser.PackResponse(response)
	if err != nil {
		return fmt.Errorf("Error packing DNS response: %w", err)
	}
	_, err = conn.WriteToUDP(binaryResponse, clientAddr)
	if err != nil {
		return fmt.Errorf("Error writing DNS response to UDP: %w", err)
	}
	return nil
}
