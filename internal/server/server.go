package server

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/Unfield/Odin-DNS/internal/config"
	"github.com/Unfield/Odin-DNS/internal/parser"
)

func StartServer(config *config.Config) {
	logger := slog.Default().WithGroup("DNS-Server")

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
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			logger.Error("Error reading from UDP", "error", err)
			continue
		}

		req, error := parser.ParseRequest(buffer)
		if error != nil {
			logger.Error("Error parsing DNS request", "error", error)
			continue
		}

		logger.Debug("Received DNS request", "client", clientAddr.String(), "request", req)

		if req.Header.Flags.QR {
			//TODO: Handle response
			logger.Warn("Received a response instead of a request")
			continue
		}

		if len(req.Questions) == 0 {
			logger.Warn("Received request with no questions")
			continue
		}

		logger.Info("Processing DNS request", "domain", req.Questions[0].Name)

		// Here you would typically look up the domain in your DNS records

		_, err = conn.WriteToUDP(buffer[:n], clientAddr)
		if err != nil {
			logger.Error("Error writing to UDP", "error", err)
			continue
		}
	}
}
