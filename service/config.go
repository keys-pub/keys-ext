package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	strings "strings"

	"github.com/pkg/errors"
)

// Config ...
type Config struct {
	appName string
	values  values
}

// Build describes build flags.
type Build struct {
	Version string
	Commit  string
	Date    string
}

func (b Build) String() string {
	return fmt.Sprintf("%s %s %s", b.Version, b.Commit, b.Date)
}

// AppName returns current app name.
func (c Config) AppName() string {
	return c.appName
}

// AppDir is where app related files are persisted.
func (c Config) AppDir() string {
	p, err := c.AppPath("", false)
	if err != nil {
		panic(err)
	}
	return p
}

// LogsDir is where logs are written.
func (c Config) LogsDir() string {
	p, err := c.LogsPath("", false)
	if err != nil {
		panic(err)
	}
	return p
}

// AppPath ...
func (c Config) AppPath(fileName string, makeDir bool) (string, error) {
	return SupportPath(c.AppName(), fileName, makeDir)
}

// LogsPath ...
func (c Config) LogsPath(fileName string, makeDir bool) (string, error) {
	return LogsPath(c.AppName(), fileName, makeDir)
}

func (c Config) certPath(makeDir bool) (string, error) {
	return c.AppPath("ca.pem", makeDir)
}

// SupportPath ...
func SupportPath(appName string, fileName string, makeDir bool) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		dir := filepath.Join(DefaultHomeDir(), "Library", "Application Support")
		return path(dir, appName, fileName, makeDir)
	case "windows":
		dir := os.Getenv("LOCALAPPDATA")
		if dir == "" {
			panic("LOCALAPPDATA not set")
		}
		return path(dir, appName, fileName, makeDir)
	case "linux":
		dir := os.Getenv("XDG_DATA_HOME")
		if dir == "" {
			dir = filepath.Join(DefaultHomeDir(), ".local", "share")
		}
		return path(dir, appName, fileName, makeDir)
	default:
		panic(fmt.Sprintf("unsupported platform %s", runtime.GOOS))
	}

}

// LogsPath ...
func LogsPath(appName string, fileName string, makeDir bool) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		dir := filepath.Join(DefaultHomeDir(), "Library", "Logs")
		return path(dir, appName, fileName, makeDir)
	case "windows":
		dir := os.Getenv("LOCALAPPDATA")
		if dir == "" {
			panic("LOCALAPPDATA not set")
		}
		return path(dir, appName, fileName, makeDir)
	case "linux":
		dir := os.Getenv("XDG_CACHE_HOME")
		if dir == "" {
			dir = filepath.Join(DefaultHomeDir(), ".cache")
		}
		return path(dir, appName, fileName, makeDir)
	default:
		panic(fmt.Sprintf("unsupported platform %s", runtime.GOOS))
	}
}

func path(dir string, appName string, fileName string, makeDir bool) (string, error) {
	if appName == "" {
		return "", errors.Errorf("appName not specified")
	}
	dir = filepath.Join(dir, appName)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logger.Infof("Creating directory: %s", dir)
		err := os.MkdirAll(dir, 0700)
		if err != nil {
			return "", err
		}
	}
	path := dir
	if fileName != "" {
		path = filepath.Join(path, fileName)
	}
	return path, nil
}

// DefaultHomeDir returns current user home directory (or "" on error).
func DefaultHomeDir() string {
	// TODO: Switch to UserHomeDir in go 1.12
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir
}

type values struct {
	Server              string      `json:"server"`
	KeyringType         KeyringType `json:"keyringType"`
	LogLevel            string      `json:"logLevel"`
	Port                int         `json:"port"`
	DisablePromptKeygen bool        `json:"disablePromptKeygen"`
	DisablePromptUser   bool        `json:"disablePromptUser"`
}

// KeyringType is the keyring to use.
type KeyringType string

const (
	// KeyringTypeDefault is the default, using the system keyring.
	KeyringTypeDefault KeyringType = ""
	// KeyringTypeFS uses the FS based keyring.
	KeyringTypeFS KeyringType = "fs"
	// KeyringTypeMem uses the in memory based keyring.
	KeyringTypeMem KeyringType = "mem"
)

// NewConfig creates a Config.
func NewConfig(appName string) (*Config, error) {
	if appName == "" {
		return nil, errors.Errorf("no app name")
	}
	cfg := &Config{
		appName: appName,
	}
	if err := cfg.Load(); err != nil {
		return nil, err
	}
	return cfg, nil
}

// Load ...
func (c *Config) Load() error {
	path, err := c.AppPath("config.json", false)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var values values
	if err := json.Unmarshal(b, &values); err != nil {
		return err
	}
	c.values = values
	return nil
}

// Save ...
func (c *Config) Save() error {
	path, err := c.AppPath("config.json", true)
	if err != nil {
		return err
	}
	b, err := json.Marshal(c.values)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, b, 0600); err != nil {
		return err
	}
	return nil
}

// Reset removes saved values.
func (c *Config) Reset() error {
	path, err := c.AppPath("config.json", true)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(path)
}

// Server ...
func (c *Config) Server() string {
	if c.values.Server == "" {
		return "https://keys.pub"
	}
	return c.values.Server
}

// SetServer ...
func (c *Config) SetServer(s string) {
	c.values.Server = s
}

// Port returns port to use to connect to service.
func (c Config) Port() int {
	if c.values.Port == 0 {
		return 10001
	}
	return c.values.Port
}

// SetPort set port.
func (c *Config) SetPort(n int) {
	c.values.Port = n
}

// LogLevel ...
func (c *Config) LogLevel() LogLevel {
	l, _ := parseLogLevel(c.values.LogLevel)
	return l
}

// SetLogLevel ...
func (c *Config) SetLogLevel(l LogLevel) {
	c.values.LogLevel = l.String()
}

// KeyringType ...
func (c Config) KeyringType() KeyringType {
	return c.values.KeyringType
}

// SetKeyringType ...
func (c *Config) SetKeyringType(t KeyringType) {
	c.values.KeyringType = t
}

// DisablePromptKeygen ...
func (c *Config) DisablePromptKeygen() bool {
	return c.values.DisablePromptKeygen
}

// SetDisablePromptKeygen ...
func (c *Config) SetDisablePromptKeygen(b bool) {
	c.values.DisablePromptKeygen = b
}

// DisablePromptUser ...
func (c *Config) DisablePromptUser() bool {
	return c.values.DisablePromptUser
}

// SetDisablePromptUser ...
func (c *Config) SetDisablePromptUser(b bool) {
	c.values.DisablePromptUser = b
}

// Export ...
func (c Config) Export() ([]byte, error) {
	return json.MarshalIndent(c.values, "", "  ")
}

// Map returns config as map values.
func (c Config) Map() map[string]string {
	return map[string]string{
		"keyringType":       string(c.values.KeyringType),
		"logLevel":          string(c.values.LogLevel),
		"port":              strconv.Itoa(c.values.Port),
		"disablePromptUser": truthyString(c.values.DisablePromptUser),
	}
}

// Set value.
func (c *Config) Set(key string, value string) error {
	switch key {
	case "keyringType":
		switch value {
		case "default":
			c.SetKeyringType(KeyringTypeDefault)
		case string(KeyringTypeFS):
			c.SetKeyringType(KeyringTypeFS)
		case string(KeyringTypeMem):
			c.SetKeyringType(KeyringTypeMem)
		default:
			return errors.Errorf("invalid value for keyringType")
		}
		return nil
	case "port":
		port, portErr := strconv.Atoi(value)
		if portErr != nil {
			return portErr
		}
		c.SetPort(port)
		return nil
	case "logLevel":
		l, ok := parseLogLevel(value)
		if !ok {
			return errors.Errorf("invalid value for logLevel")
		}
		c.SetLogLevel(l)
		return nil
	case "disablePromptUser":
		if value == "" {
			return errors.Errorf("empty value")
		}
		b, err := truthy(value)
		if err != nil {
			return err
		}
		c.SetDisablePromptUser(b)
		return nil
	default:
		return errors.Errorf("unknown config key")
	}
}

// Config (RPC) ...
func (s *service) Config(ctx context.Context, req *ConfigRequest) (*ConfigResponse, error) {
	return &ConfigResponse{
		Config: s.cfg.Map(),
	}, nil
}

// ConfigSet (RPC) ...
func (s *service) ConfigSet(ctx context.Context, req *ConfigSetRequest) (*ConfigSetResponse, error) {
	if err := s.cfg.Set(req.Key, req.Value); err != nil {
		return nil, err
	}
	if err := s.cfg.Save(); err != nil {
		return nil, err
	}
	return &ConfigSetResponse{}, nil
}

func truthy(s string) (bool, error) {
	s = strings.TrimSpace(s)
	switch s {
	case "1", "t", "true", "y", "yes":
		return true, nil
	case "0", "f", "false", "n", "no":
		return false, nil
	default:
		return false, errors.Errorf("invalid value: %s", s)
	}
}

func truthyString(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
