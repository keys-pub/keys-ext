package service

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"

	middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"github.com/keys-pub/keys"
	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/keys-pub/keys-ext/http/client"
	"github.com/keys-pub/keys-ext/vault"
	"github.com/keys-pub/keys-ext/wormhole"
	"github.com/keys-pub/keys-ext/wormhole/sctp"
	"github.com/keys-pub/keys/link"
	"github.com/keys-pub/keys/request"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/user"
	"github.com/mercari/go-grpc-interceptor/panichandler"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func newProtoService(cfg *Config, build Build, auth *auth) (*service, error) {
	req := request.NewHTTPRequestor()
	srv, err := newService(cfg, build, auth, req, tsutil.NewClock())
	if err != nil {
		return nil, err
	}
	return srv, nil
}

func setupLogging(cfg *Config, logPath string) (Logger, LogInterceptor) {
	return setupLogrus(cfg, logPath)
}

func logFatal(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

// Run the service.
func Run(build Build) {
	appName := flag.String("app", "Keys", "app name")
	logPath := flag.String("log-path", "", "log path")
	version := flag.Bool("version", false, "print version")
	port := flag.Int("port", 0, "port to listen")
	logLevel := flag.String("log-level", "", "log level")

	flag.Parse()

	if *version {
		fmt.Printf("%s\n", build)
		return
	}

	cfg, err := NewConfig(*appName)
	if err != nil {
		logFatal(errors.Wrapf(err, "failed to load config"))
	}

	if len(flag.Args()) > 0 {
		logFatal(errors.Errorf("Invalid arguments. Did you mean to run `keys`?"))
	}

	// Set config from flags
	if err := cfg.saveLogLevelFlag(*logLevel); err != nil {
		logFatal(err)
	}
	if err := cfg.savePortFlag(*port); err != nil {
		logFatal(err)
	}

	// TODO: Disable logging by default

	lg, lgi := setupLogging(cfg, *logPath)
	SetLogger(lg)
	keys.SetLogger(lg)
	user.SetLogger(lg)
	link.SetLogger(lg)
	saltpack.SetLogger(lg)
	vault.SetLogger(lg)
	client.SetLogger(lg)
	wormhole.SetLogger(lg)
	sctp.SetLogger(lg)

	logger.Debugf("Running %v", os.Args)

	if err := checkSupportedOS(); err != nil {
		logFatal(err)
	}
	if runtime.GOOS == "darwin" {
		if err := checkCodesigned(); err != nil {
			logFatal(err)
		}
	}

	logger.Infof("Version: %s", build)
	logger.Infof("Log level: %s", cfg.LogLevel().String())

	panichandler.InstallPanicHandler(func(ctx context.Context, r interface{}) {
		logrus.Errorf("Panic: %v; %s", r, string(debug.Stack()))
	})

	if err := runService(cfg, build, lgi); err != nil {
		logFatal(err)
	}
}

// ServeFn starts the service.
type ServeFn func() error

// CloseFn closes the service.
type CloseFn func()

// TODO: Protect against incompatible downgrades

func runService(cfg *Config, build Build, lgi LogInterceptor) error {
	if IsPortInUse(cfg.Port()) {
		return errors.Errorf("port %d in use; is keysd already running?", cfg.Port())
	}

	cert, err := GenerateCertificate(cfg, true)
	if err != nil {
		return err
	}
	defer func() { _ = DeleteCertificate(cfg) }()

	serveFn, closeFn, serveErr := NewServiceFn(cfg, build, cert, lgi)
	if serveErr != nil {
		return serveErr
	}
	defer closeFn()
	return serveFn()
}

// NewServiceFn ...
func NewServiceFn(cfg *Config, build Build, cert *keys.CertificateKey, lgi LogInterceptor) (ServeFn, CloseFn, error) {
	var opts []grpc.ServerOption

	if cert == nil {
		return nil, nil, errNoCertFound{}
	}
	tlsCert := cert.TLSCertificate()
	creds := credentials.NewServerTLSFromCert(&tlsCert)

	opts = []grpc.ServerOption{
		grpc.Creds(creds),
	}

	auth := newAuth(cfg)

	lgi.Replace()

	opts = append(opts, middleware.WithUnaryServerChain(
		ctxtags.UnaryServerInterceptor(ctxtags.WithFieldExtractor(ctxtags.CodeGenRequestFieldExtractor)),
		lgi.Unary(),
		auth.unaryInterceptor,
		panichandler.UnaryServerInterceptor,
	),
		middleware.WithStreamServerChain(
			ctxtags.StreamServerInterceptor(ctxtags.WithFieldExtractor(ctxtags.CodeGenRequestFieldExtractor)),
			lgi.Stream(),
			auth.streamInterceptor,
			panichandler.StreamServerInterceptor,
		),
	)
	grpcServer := grpc.NewServer(opts...)

	service, err := newProtoService(cfg, build, auth)
	if err != nil {
		return nil, nil, err
	}

	if err := service.Open(); err != nil {
		return nil, nil, err
	}

	// Keys service
	logger.Infof("Registering Keys service...")
	RegisterKeysServer(grpcServer, service)

	// FIDO2
	fido2Plugin, err := fido2.OpenPlugin(filepath.Join(exeDir(), "fido2.so"))
	if err != nil {
		logger.Errorf("fido2 plugin is not available: %v", err)
	} else {
		logger.Infof("Registering FIDO2 plugin...")
		fido2.RegisterAuthServer(grpcServer, fido2Plugin)
		auth.fas = fido2Plugin
	}

	logger.Infof("Listening for connections on port %d", cfg.Port())
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.Port()))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to tcp listen")
	}

	serveFn := func() error {
		if err := writePID(cfg); err != nil {
			return err
		}
		return grpcServer.Serve(lis)
	}
	closeFn := func() {
		grpcServer.Stop()
		service.Close()
	}
	return serveFn, closeFn, nil
}

// IsPortInUse returns true if port is currently in use.
func IsPortInUse(port int) bool {
	lis, lisErr := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if lisErr != nil {
		return true
	}
	_ = lis.Close()
	return false
}

func writePID(cfg *Config) error {
	path, err := cfg.AppPath("pid", false)
	if err != nil {
		return err
	}
	pid := os.Getpid()
	return ioutil.WriteFile(path, []byte(strconv.Itoa(pid)), filePerms)
}
