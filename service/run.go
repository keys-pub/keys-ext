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
	wsclient "github.com/keys-pub/keys-ext/ws/client"
	"github.com/keys-pub/keys/http"
	"github.com/keys-pub/keys/saltpack"
	"github.com/keys-pub/keys/tsutil"
	"github.com/keys-pub/keys/user"
	"github.com/keys-pub/keys/user/services"
	"github.com/mercari/go-grpc-interceptor/panichandler"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func setupLogging(logLevel LogLevel, logPath string) (*logrus.Logger, LogInterceptor) {
	return setupLogrus(logLevel, logPath)
}

func logFatal(err error) {
	fmt.Fprintf(os.Stderr, "%v\n", err)
	os.Exit(1)
}

type args struct {
	appName  *string
	logPath  *string
	version  *bool
	port     *int
	logLevel *string
}

// Run the service.
func Run(build Build) {
	args := args{}
	args.appName = flag.String("app", "Keys", "app name")
	args.logPath = flag.String("log-path", "", "log path")
	args.version = flag.Bool("version", false, "print version")
	args.port = flag.Int("port", 0, "port to listen")
	args.logLevel = flag.String("log-level", "", "log level")

	flag.Parse()

	if *args.version {
		fmt.Printf("%s\n", build)
		return
	}

	logLevel, ok := parseLogLevel(*args.logLevel)
	if !ok {
		logFatal(errors.Errorf("invalid log level"))
	}

	env, err := NewEnv(*args.appName)
	if err != nil {
		logFatal(errors.Wrapf(err, "failed to load config"))
	}

	if len(flag.Args()) > 0 {
		logFatal(errors.Errorf("Invalid arguments. Did you mean to run `keys`?"))
	}

	// Save env
	if err := env.savePortFlag(*args.port); err != nil {
		logFatal(err)
	}

	// TODO: Disable logging by default?
	// TODO: Include package name in logging

	lg, lgi := setupLogging(logLevel, *args.logPath)
	SetLogger(lg)
	client.SetLogger(newPackageLogger(lg, "keys-ext/http/client"))
	keys.SetLogger(newPackageLogger(lg, "keys"))
	saltpack.SetLogger(newPackageLogger(lg, "keys/saltpack"))
	sctp.SetLogger(newPackageLogger(lg, "keys-ext/wormhole/sctp"))
	user.SetLogger(newPackageLogger(lg, "keys/user"))
	services.SetLogger(newPackageLogger(lg, "keys/user/services"))
	vault.SetLogger(newPackageLogger(lg, "keys-ext/vault"))
	wormhole.SetLogger(newPackageLogger(lg, "keys-ext/wormhole"))
	wsclient.SetLogger(newPackageLogger(lg, "keys-ext/ws/client"))

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
	logger.Infof("Log level: %s", logLevel.String())

	panichandler.InstallPanicHandler(func(ctx context.Context, r interface{}) {
		logrus.Errorf("Panic: %v; %s", r, string(debug.Stack()))
	})

	if err := runService(env, build, lgi); err != nil {
		logFatal(err)
	}
}

// ServeFn starts the service.
type ServeFn func() error

// CloseFn closes the service.
type CloseFn func()

// TODO: Protect against incompatible downgrades

func runService(env *Env, build Build, lgi LogInterceptor) error {
	if IsPortInUse(env.Port()) {
		return errors.Errorf("port %d in use; is keysd already running?", env.Port())
	}

	cert, err := GenerateCertificate(env, true)
	if err != nil {
		return err
	}
	defer func() { _ = DeleteCertificate(env) }()

	serveFn, closeFn, serveErr := NewServiceFn(env, build, cert, lgi)
	if serveErr != nil {
		return serveErr
	}
	defer closeFn()
	return serveFn()
}

// NewServiceFn ...
func NewServiceFn(env *Env, build Build, cert *keys.CertificateKey, lgi LogInterceptor) (ServeFn, CloseFn, error) {
	var opts []grpc.ServerOption

	if cert == nil {
		return nil, nil, errNoCertFound{}
	}
	tlsCert := cert.TLSCertificate()
	creds := credentials.NewServerTLSFromCert(&tlsCert)

	opts = []grpc.ServerOption{
		grpc.Creds(creds),
	}

	auth := newAuth(env)

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

	client := http.NewClient()
	service, err := newService(env, build, auth, client, tsutil.NewClock())
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
		fido2.RegisterFIDO2Server(grpcServer, fido2Plugin)
		auth.fas = fido2Plugin
	}

	logger.Infof("Listening for connections on port %d", env.Port())
	lis, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", env.Port()))
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to tcp listen")
	}

	serveFn := func() error {
		if err := writePID(env); err != nil {
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

func writePID(env *Env) error {
	path, err := env.AppPath("pid", false)
	if err != nil {
		return err
	}
	pid := os.Getpid()
	return ioutil.WriteFile(path, []byte(strconv.Itoa(pid)), filePerms)
}
