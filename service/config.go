package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/keys-pub/keys/env"
	"github.com/pkg/errors"
)

// Config for app runtime.
// Do not store anything sensitive in here, values are saved clear and can be
// modified at will.
// Config is not authenticated.
type Config struct {
	appName string
	values  map[string]string
}

// NewConfig loads a Config.
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

// Config key names
const serverCfgKey = "server"
const portCfgKey = "port"
const logLevelCfgKey = "logLevel"

var configKeys = []string{serverCfgKey, portCfgKey, logLevelCfgKey}

// IsKey returns true if config key is recognized.
func (c Config) IsKey(s string) bool {
	for _, k := range configKeys {
		if s == k {
			return true
		}
	}
	return false
}

// Port to connect.
func (c Config) Port() int {
	return c.GetInt(portCfgKey, 22405)
}

// Server to connect to.
func (c Config) Server() string {
	return c.Get(serverCfgKey, "https://keys.pub")
}

// LogLevel for logging.
func (c *Config) LogLevel() LogLevel {
	ll := c.Get(logLevelCfgKey, "")
	l, _ := parseLogLevel(ll)
	return l
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
func (c Config) AppPath(file string, makeDir bool) (string, error) {
	opts := []env.PathOption{env.Dir(c.AppName()), env.File(file)}
	if makeDir {
		opts = append(opts, env.Mkdir())
	}
	return env.AppPath(opts...)
}

// LogsPath ...
func (c Config) LogsPath(file string, makeDir bool) (string, error) {
	opts := []env.PathOption{env.Dir(c.AppName()), env.File(file)}
	if makeDir {
		opts = append(opts, env.Mkdir())
	}
	return env.LogsPath(opts...)
}

func (c Config) certPath(makeDir bool) (string, error) {
	return c.AppPath("ca.pem", makeDir)
}

// Path to config file.
func (c *Config) Path(makeDir bool) (string, error) {
	return c.AppPath("config.json", makeDir)
}

// Load ...
func (c *Config) Load() error {
	path, err := c.Path(false)
	if err != nil {
		return err
	}

	var values map[string]string

	exists, err := pathExists(path)
	if err != nil {
		return err
	}
	if exists {
		b, err := ioutil.ReadFile(path) // #nosec
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &values); err != nil {
			return err
		}
	}
	if values == nil {
		values = map[string]string{}
	}
	c.values = values
	return nil
}

// Save ...
func (c *Config) Save() error {
	path, err := c.Path(true)
	if err != nil {
		return err
	}
	b, err := json.Marshal(c.values)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(path, b, filePerms); err != nil {
		return err
	}
	return nil
}

// Reset removes saved values.
func (c *Config) Reset() error {
	path, err := c.Path(false)
	if err != nil {
		return err
	}

	exists, err := pathExists(path)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	return os.Remove(path)
}

// Export ...
func (c Config) Export() ([]byte, error) {
	return json.MarshalIndent(c.values, "", "  ")
}

// Get config value.
func (c *Config) Get(key string, dflt string) string {
	v, ok := c.values[key]
	if !ok {
		return dflt
	}
	return v
}

// GetInt gets config value as int.
func (c *Config) GetInt(key string, dflt int) int {
	v, ok := c.values[key]
	if !ok {
		return dflt
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		logger.Warningf("config value %s not an int", key)
		return 0
	}
	return n

}

// GetBool gets config value as bool.
func (c *Config) GetBool(key string) bool {
	v, ok := c.values[key]
	if !ok {
		return false
	}
	b, _ := truthy(v)
	return b
}

// SetBool sets bool value for key.
func (c *Config) SetBool(key string, b bool) {
	c.Set(key, truthyString(b))
}

// SetInt sets int value for key.
func (c *Config) SetInt(key string, n int) {
	c.Set(key, strconv.Itoa(n))
}

// Set value.
func (c *Config) Set(key string, value string) {
	c.values[key] = value
}

func (c *Config) saveLogLevelFlag(s string) error {
	if s == "" {
		return nil
	}
	_, ok := parseLogLevel(s)
	if !ok {
		return errors.Errorf("invalid log-level")
	}
	c.Set(logLevelCfgKey, s)
	return c.Save()
}

func (c *Config) savePortFlag(port int) error {
	if port == 0 {
		return nil
	}
	c.SetInt(portCfgKey, port)
	return c.Save()
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
