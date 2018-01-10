package main

import (
	"crypto/tls"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/deciphernow/gm-fabric-go/tlsutil"

    pb "{{.PBImport}}"
)

func main() {
    os.Exit(run())
}

func run() int {
    var grpcServerAddress string
	var testCertDir string
    var client pb.{{.GoServiceName}}Client
    var err error

    logger := zerolog.New(os.Stderr).With().Timestamp().Logger().
		Output(zerolog.ConsoleWriter{Out: os.Stderr})

	pflag.StringVar(
		&grpcServerAddress,
		"address",
		"",
		"address of grpc server",
	)
	pflag.StringVar(
		&testCertDir,
		"test-cert-dir",
		"",
		"(if TLS) directory holding test certificates",
	)
	pflag.Parse()
    if grpcServerAddress == "" {
        logger.Error().Msg("You must specify server address. (--address)")
        return 1
    }

    logger.Info().Str("grpc-client", "{{.ServiceName}}").
    Str("address", grpcServerAddress).
    Str("test-certs", testCertDir).Msg("starting")

    if client, err = newClient(grpcServerAddress, testCertDir); err != nil {
        logger.Error().AnErr("newClient", err).Msg("")
        return 1
	}

    if err = runTest(logger, client); err != nil {
        logger.Error().AnErr("runTest", err).Msg("")
        return 1
	}

    logger.Info().Str("grpc client", "{{.ServiceName}}").Msg("terminating normally")
    return 0
}

func newClient(
	serverAddress string,
	testCertDir string,
) (pb.TestServiceClient, error) {

	var opts []grpc.DialOption
	var conn *grpc.ClientConn
	var err error

	if testCertDir == "" {
		opts = append(opts, grpc.WithInsecure())
	} else {
		var tlsConf *tls.Config

		tlsConf, err := tlsutil.NewTLSClientConfig(
			filepath.Join(testCertDir, "root.crt"),                      // ca_cert_path
			filepath.Join(testCertDir, "server.localdomain.chain.crt"),  // server_cert_path
			filepath.Join(testCertDir, "server.localdomain.nopass.key"), // server_key_path
			"server.localdomain",                                        // server_cert_name
		)
		if err != nil {
			return nil, errors.Wrap(err, "tlsutil.NewTLSClientConfig")
		}

		creds := credentials.NewTLS(tlsConf)
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	conn, err = grpc.Dial(serverAddress, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "grpc.Dial(%s", serverAddress)
	}

	return pb.NewTestServiceClient(conn), nil
}
