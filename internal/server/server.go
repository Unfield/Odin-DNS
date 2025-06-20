package server

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/Unfield/Odin-DNS/internal/config"
	"github.com/Unfield/Odin-DNS/internal/parser"
)

func StartServer(config *config.Config) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", config.DNS_PORT))
	if err != nil {
		slog.Error("Error resolving address", "port", config.DNS_PORT, "error", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		slog.Error("Error listening on UDP port", "port", config.DNS_PORT, "error", err)
		return
	}
	defer conn.Close()

	slog.Info("Odin DNS server is running", "port", addr.Port)

	buffer := make([]byte, config.BUFFER_SIZE)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			slog.Error("Error reading from UDP", "error", err)
			continue
		}

		req, error := parser.ParseRequest(buffer)
		if error != nil {
			slog.Error("Error parsing DNS request", "error", error)
			continue
		}

		slog.Debug("Received DNS request", "client", clientAddr.String(), "request", req)

		if req.Header.Flags.QR {
			//TODO: Handle response
			slog.Warn("Received a response instead of a request")
			continue
		}

		if len(req.Questions) == 0 {
			slog.Warn("Received request with no questions")
			continue
		}

		slog.Info("Processing DNS request", "domain", req.Questions[0].Name)

		// Here you would typically look up the domain in your DNS records

		_, err = conn.WriteToUDP(buffer[:n], clientAddr)
		if err != nil {
			slog.Error("Error writing to UDP", "error", err)
			continue
		}
	}
}
