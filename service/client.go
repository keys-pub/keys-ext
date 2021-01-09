package service

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/keys-pub/keys-ext/auth/fido2"
	"github.com/keys-pub/keys/dstore"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
)

// Client defines the RPC client.
type Client struct {
	sync.Mutex
	keysClient  RPCClient
	fido2Client fido2.FIDO2Client
	conn        *grpc.ClientConn
	env         *Env
	connectFn   ClientConnectFn

	out io.Writer
}

// VersionDev is default for dev environment.
const VersionDev = "0.0.0-dev"

// ClientConnectFn describes client connect.
type ClientConnectFn func(env *Env, authToken string) (*grpc.ClientConn, error)

// NewClient constructs a client.
func NewClient() *Client {
	return &Client{
		connectFn: connectLocal,
		out:       os.Stderr,
	}
}

// Connect ...
func (c *Client) Connect(env *Env, authToken string) error {
	c.Lock()
	defer c.Unlock()

	if c.conn != nil {
		if err := c.Close(); err != nil {
			logger.Warningf("Error closing existing connection: %s", err)
		}
	}

	c.env = env
	conn, err := c.connectFn(env, authToken)
	if err != nil {
		return err
	}
	c.conn = conn
	c.keysClient = NewRPCClient(conn)
	c.fido2Client = fido2.NewFIDO2Client(conn)
	return nil
}

func connectLocal(env *Env, authToken string) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	certPEM, err := loadCertificate(env)
	if err != nil {
		return nil, err
	}
	if certPEM == "" {
		return nil, errNoCertFound{}
	}
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM([]byte(certPEM)) {
		return nil, errors.Errorf("failed to add cert to pool")
	}
	creds := credentials.NewClientTLSFromCert(certPool, "localhost")

	opts = append(opts, grpc.WithTransportCredentials(creds))
	opts = append(opts, grpc.WithPerRPCCredentials(newClientAuth(authToken)))
	addr := fmt.Sprintf("127.0.0.1:%d", env.Port())
	logger.Infof("Opening connection: %s", addr)
	return grpc.Dial(addr, opts...)
}

// RPCClient returns Keys RPC client.
func (c *Client) RPCClient() RPCClient {
	return c.keysClient
}

// FIDO2Client returns FIDO2 Authenticators RPC client.
func (c *Client) FIDO2Client() fido2.FIDO2Client {
	return c.fido2Client
}

// Close ...
func (c *Client) Close() error {
	var err error
	if c.conn != nil {
		err = c.conn.Close()
		c.conn = nil
	}
	c.keysClient = nil
	c.fido2Client = nil
	return err
}

// RunClient runs the command line client
func RunClient(build Build) {
	if err := checkSupportedOS(); err != nil {
		logger.Fatalf("%s", err)
	}
	if runtime.GOOS == "darwin" {
		if err := checkCodesigned(); err != nil {
			logger.Fatalf("%s", err)
		}
	}

	logger.Debugf("Running %v", os.Args)
	client := NewClient()
	defer client.Close()
	runClient(build, os.Args, client, clientFatal)
}

func newClientEnv(c *cli.Context) (*Env, error) {
	appName := c.GlobalString("app")
	if appName == "" {
		return nil, errors.Errorf("no app name")
	}
	return NewEnv(appName)
}

func runClient(build Build, args []string, client *Client, errorFn func(err error)) {
	app := cli.NewApp()
	app.Name = "keys"
	app.Version = build.String()
	app.Usage = "Cryptographic key management, signing and encryption."

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "log-level",
			Value: "warn",
			Usage: "log level (debug, info, warn, err)",
		},
		cli.StringFlag{
			Name:  "app",
			Value: "Keys",
			Usage: "app name",
		},
	}

	logger := logrus.StandardLogger()
	formatter := &logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: time.RFC3339Nano,
	}
	logger.SetFormatter(formatter)
	// logger.SetReportCaller(true)
	SetLogger(logger)

	cmds := []cli.Command{}
	cmds = append(cmds, startCommands()...)
	cmds = append(cmds, authCommands(client)...)
	cmds = append(cmds, signCommands(client)...)
	cmds = append(cmds, verifyCommands(client)...)
	cmds = append(cmds, sigchainCommands(client)...)
	cmds = append(cmds, encryptCommands(client)...)
	cmds = append(cmds, pullCommands(client)...)
	cmds = append(cmds, importCommands(client)...)
	cmds = append(cmds, exportCommands(client)...)
	cmds = append(cmds, dbCommands(client)...)
	cmds = append(cmds, otherCommands(client)...)
	cmds = append(cmds, userCommands(client)...)
	cmds = append(cmds, keyCommands(client)...)
	cmds = append(cmds, envCommands(client)...)
	cmds = append(cmds, logCommands(client)...)
	cmds = append(cmds, wormholeCommands(client)...)
	cmds = append(cmds, fido2Commands(client)...)
	cmds = append(cmds, adminCommands(client)...)
	cmds = append(cmds, vaultCommands(client)...)
	cmds = append(cmds, messageCommands(client)...)

	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name < cmds[j].Name
	})

	app.Commands = cmds

	app.Before = func(c *cli.Context) error {
		logLevel, err := logrusLevel(c.GlobalString("log-level"))
		if err != nil {
			errorFn(err)
			return err
		}
		logger.SetLevel(logLevel)
		logger.Infof("Version: %s", build.String())
		logger.Debugf("PID: %d", os.Getpid())
		logger.Debugf("UID: %d", os.Getuid())
		logger.Debugf("OS: %s", runtime.GOOS)

		env, err := newClientEnv(c)
		if err != nil {
			errorFn(err)
			return err
		}

		command := c.Args().Get(0)
		logger.Debugf("Command: %s", command)

		// Start commands don't connect to the service.
		skip := dstore.NewStringSet("", "log", "uninstall", "restart", "start", "stop", "env")
		if skip.Contains(command) {
			return nil
		}

		if build.Version != VersionDev {
			if err := autostart(env); err != nil {
				errorFn(err)
				return err
			}
		}

		authToken := os.Getenv("KEYS_AUTH")

		if err := connect(env, client, build, authToken, true); err != nil {
			errorFn(err)
			return err
		}

		return nil
	}

	if err := app.Run(args); err != nil {
		errorFn(err)
	}
}

func connect(env *Env, client *Client, build Build, authToken string, reconnect bool) error {
	logger.Debugf("Client connect...")
	if err := client.Connect(env, authToken); err != nil {
		return err
	}

	logger.Debugf("Service status...")
	status, err := client.RPCClient().RuntimeStatus(context.TODO(), &RuntimeStatusRequest{})
	if err != nil {
		return err
	}

	// TODO: Does this check happen during auth?
	if env.AppName() != status.AppName {
		return errServiceRuntime{Reason: fmt.Sprintf("service and client have different app names %s != %s", env.AppName(), status.AppName)}
	}

	if build.Version == VersionDev {
		return nil
	}

	// Check service and client running from same directories.
	exe, exeErr := executablePath()
	if exeErr != nil {
		return errors.Wrapf(exeErr, "failed to get executable path")
	}
	if status.Exe == "" {
		return errServiceRuntime{Reason: "service is running from a non-existent location"}
	}

	// Check service and client running same version.
	// If not, try to restart (if supported) and retry.
	if status.Version != build.Version {
		logger.Infof("Service client version mismatch, %s != %s", status.Version, build.Version)
		if reconnect {
			fmt.Fprintf(os.Stderr, "The keysd (service) is a different version than the client, restarting it...\n")
			// Try to restart
			if err := restart(env); err != nil {
				return errServiceRuntime{Reason: err.Error()}
			}
			logger.Infof("Reconnecting...")
			return connect(env, client, build, authToken, false)
		}

		return errDifferentVersions{VersionService: status.Version, VersionClient: build.Version}
	}

	dir := filepath.Dir(exe)
	serviceDir := filepath.Dir(status.Exe)
	if dir != serviceDir {
		return errServiceRuntime{Reason: fmt.Sprintf("service and client are running from different directories, %s != %s", serviceDir, dir)}
	}

	return nil
}

type errNoCertFound struct{}

func (e errNoCertFound) Error() string {
	return "no certificate was found"
}

type errDifferentVersions struct {
	VersionService string
	VersionClient  string
}

type errServiceRuntime struct {
	Reason string
}

func (e errServiceRuntime) Error() string {
	return e.Reason
}

func (e errDifferentVersions) Error() string {
	return fmt.Sprintf("service and client version are different, %s != %s", e.VersionService, e.VersionClient)
}

func clientFatal(err error) {
	// TODO: Use executable name instead of `keys`.
	switch err := err.(type) {
	case errDifferentVersions:
		fmt.Fprintf(os.Stderr, "The service and client are running different versions, %s != %s.\n", err.VersionService, err.VersionClient)
		fmt.Fprintf(os.Stderr, "You may need to `keys restart`.\n")
		os.Exit(1)
	case errServiceRuntime:
		fmt.Fprintf(os.Stderr, "The service had a runtime error: %s.\n", err.Reason)
		fmt.Fprintf(os.Stderr, "You may need to `keys restart`.\n")
		os.Exit(1)
	}

	st, ok := status.FromError(err)
	if !ok {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	logger.Infof("Received error %d %s", st.Code(), st.Message())

	switch st.Code() {
	case codes.Unavailable:
		fmt.Fprintf(os.Stderr, "Service is unavailable, run `keys start`.\n")
	case codes.PermissionDenied:
		fmt.Fprintf(os.Stderr, "Permission denied.\n")
	case codes.Unauthenticated:
		switch st.Message() {
		case "invalid password":
			fmt.Fprintf(os.Stderr, "Invalid password.\n")
		case "invalid auth":
			fmt.Fprintf(os.Stderr, "Invalid auth.\n")
		default:
			fmt.Fprintf(os.Stderr, "Authorization required, run `keys auth`.\n")
		}
	case codes.Unknown:
		// TODO: Use error codes from service for nicer error messages
		fmt.Fprintf(os.Stderr, "%s\n", st.Message())
	default:
		fmt.Fprintf(os.Stderr, "%s (%d)\n", st.Message(), st.Code())
	}

	exitCode := int(st.Code())
	os.Exit(exitCode)
}

func logrusLevel(s string) (logrus.Level, error) {
	switch s {
	case "debug":
		return logrus.DebugLevel, nil
	case "info":
		return logrus.InfoLevel, nil
	case "warn":
		return logrus.WarnLevel, nil
	case "err":
		return logrus.ErrorLevel, nil
	default:
		return logrus.DebugLevel, errors.Errorf("log should one of: debug, info, warn, err")
	}
}

func logrusFromLevel(l LogLevel) logrus.Level {
	switch l {
	case DebugLevel:
		return logrus.DebugLevel
	case InfoLevel:
		return logrus.InfoLevel
	case WarnLevel:
		return logrus.WarnLevel
	case ErrLevel:
		return logrus.ErrorLevel
	default:
		return logrus.DebugLevel
	}
}
