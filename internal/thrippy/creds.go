package thrippy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log/slog"
	"os"

	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/tzrikka/timpani/internal/logger"
)

// SecureCreds initializes gRPC client credentials using TLS or mTLS,
// based on CLI flags. Errors abort the application with a log message.
// If the flag "--dev" is specified, we use insecure credentials instead.
func SecureCreds(ctx context.Context, cmd *cli.Command) credentials.TransportCredentials {
	if cmd.Bool("dev") {
		return insecure.NewCredentials()
	}

	// Both TLS and mTLS.
	caPath := cmd.String("thrippy-server-ca-cert")
	nameOverride := cmd.String("thrippy-server-name-override")
	// Only mTLS.
	certPath := cmd.String("thrippy-client-cert")
	keyPath := cmd.String("thrippy-client-key")

	// The server's CA cert is required either way (on many Linux systems,
	// "/etc/ssl/cert.pem" contains the system-wide set of root CAs).
	if caPath == "" {
		logger.Fatal(ctx, "missing server CA cert file for gRPC client with m/TLS")
	}

	// Using mTLS requires the client's X.509 PEM-encoded public cert
	// and private key. If one of them is missing it's an error.
	if certPath == "" && keyPath != "" {
		logger.Fatal(ctx, "missing client public cert file for gRPC client with mTLS")
	}
	if certPath != "" && keyPath == "" {
		logger.Fatal(ctx, "missing client private key file for gRPC client with mTLS")
	}

	// If both of them are missing, we use TLS.
	if certPath == "" && keyPath == "" {
		return newClientTLSFromFile(ctx, caPath, nameOverride, nil)
	}

	// If all 3 are specified, we use mTLS.
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		logger.FatalErrorContext(ctx, "failed to load client PEM key pair for gRPC client with mTLS",
			err, slog.String("cert", certPath), slog.String("key", keyPath))
	}

	return newClientTLSFromFile(ctx, caPath, nameOverride, []tls.Certificate{cert})
}

// newClientTLSFromFile constructs TLS credentials from the provided root
// certificate authority certificate file(s) to validate server connections.
//
// This function is based on [credentials.NewClientTLSFromFile], but uses
// TLS 1.3 as the minimum version (instead of 1.2), and support mTLS too.
func newClientTLSFromFile(ctx context.Context, caPath, serverNameOverride string, certs []tls.Certificate) credentials.TransportCredentials {
	b, err := os.ReadFile(caPath) //gosec:disable G304 // Specified by admin by design.
	if err != nil {
		logger.FatalErrorContext(ctx, "failed to read server CA cert file for gRPC client",
			err, slog.String("path", caPath))
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		logger.Fatal(ctx, "failed to parse server CA cert file for gRPC client", slog.String("path", caPath))
	}

	cfg := &tls.Config{
		RootCAs:    cp,
		ServerName: serverNameOverride,
		MinVersion: tls.VersionTLS13,
	}
	if len(certs) > 0 {
		cfg.Certificates = certs
	}

	return credentials.NewTLS(cfg)
}
