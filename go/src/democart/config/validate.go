package config

import (
	"flag"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/hcl"
	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"
)

var (
	configFilesFlag = flag.String("configs", "", "config file paths as csv")

	// flag overrides
	dbURLFlag        = flag.String("db_url", "", "database url")
	logLevelFlag     = flag.String("loglevel", "", "log level")
	publicAPIURLFlag = flag.String("public_api_url", "", "public api url")
	publicIDPURLFlag = flag.String("public_idp_url", "", "public idp url")
	clientHostsFlag  = flag.String("client_hosts", "", "csv client hosts")

	// database env var overrides
	dbDriverEnv = os.Getenv("DATABASE_DRIVER")
	dbHostEnv   = os.Getenv("DATABASE_HOST")
	dbPortEnv   = os.Getenv("DATABASE_PORT")
	dbSSLEnv    = os.Getenv("DATABASE_SSLMODE")
	dbUserEnv   = os.Getenv("DATABASE_USER")
	dbPassEnv   = os.Getenv("DATABASE_PASSWORD")
	dbDBNameEnv = os.Getenv("DATABASE_NAME")

	// services env var overrides
	publicAPIURLEnv = os.Getenv("PUBLIC_API_URL")
	publicIDPURLEnv = os.Getenv("PUBLIC_IDP_URL")
	clientHostsEnv  = os.Getenv("CLIENT_HOSTS")

	configErr = errs.Class("configuration")
)

type Configs struct {
	Version    string
	DBURL      *url.URL
	APISlug    string
	APIAddress string
	IDPAddress string
	//MetricAddress           string
	GracefulShutdownTimeout time.Duration
	WriteTimeout            time.Duration
	ReadTimeout             time.Duration
	IdleTimeout             time.Duration
	IDPPasswordSalt         string
	IDPClientID             string
	IDPClientSecret         string
	LogLevel                logrus.Level
	DeveloperMode           bool
	InsecureRequestsMode    bool
	ClientHosts             []*url.URL
	PublicAPIURL            *url.URL
	PublicIDPURL            *url.URL
}

func (c *Configs) SetVersion(v string) { c.Version = v }

// Parse will set the configuration values pulled from the provided config
// file, any config flags, and any environment variables. env vars have highest
// priority, then config flags, then the config file. If any values overlap and
// *DON'T* match, an error is thrown.
func Parse() (*Configs, error) {
	flag.Parse()

	raw := rawConfigs{}

	// set the configurations specified by the config filenames
	err := raw.setConfigFiles()
	if err != nil {
		return nil, err
	}

	// set the configurations specified by the flags
	err = raw.setNoChangeConfigFlags()
	if err != nil {
		return nil, err
	}

	// set the configurations specified by the environment variables
	err = raw.setNoChangeEnvVars()
	if err != nil {
		return nil, err
	}

	return raw.validate()
}

type rawConfigs struct {
	DBURL      string `hcl:"db_url"`
	APISlug    string `hcl:"api_slug"`
	APIAddress string `hcl:"api_addr"`
	IDPAddress string `hcl:"idp_addr"`
	//MetricAddress           string   `hcl:"metric_addr"`
	GracefulShutdownTimeout int      `hcl:"graceful_shutdown_timeout_sec"`
	WriteTimeout            int      `hcl:"write_timeout_sec"`
	ReadTimeout             int      `hcl:"read_timeout_sec"`
	IdleTimeout             int      `hcl:"idle_timeout_sec"`
	IDPPasswordSalt         string   `hcl:"idp_password_salt"`
	IDPClientID             string   `hcl:"idp_client_id"`
	IDPClientSecret         string   `hcl:"idp_client_secret"`
	LogLevel                string   `hcl:"loglevel"`
	DeveloperMode           bool     `hcl:"developer_mode"`
	InsecureRequestsMode    bool     `hcl:"insecure_requests_mode"`
	ClientHosts             []string `hcl:"client_hosts"`
	PublicAPIURL            string   `hcl:"public_api_url"`
	PublicIDPURL            string   `hcl:"public_idp_url"`
}

// setConfigFiles will set all of the values provided in the config files,
// stomping over any existing values
func (raw *rawConfigs) setConfigFiles() error {
	if configFilesFlag == nil || *configFilesFlag == "" {
		return nil
	}

	configFiles := strings.Split(*configFilesFlag, ",")
	for _, configFile := range configFiles {
		hclBytes, err := ioutil.ReadFile(configFile)
		if err != nil {
			return configErr.Wrap(err)
		}

		if err := hcl.Unmarshal(hclBytes, raw); err != nil {
			return configErr.Wrap(err)
		}
	}

	return nil
}

// setNoChangeConfigFlags will set all of the values provided as config flags
// as long as they don't change existing non-empty values in Configs
func (raw *rawConfigs) setNoChangeConfigFlags() error {
	err := setStringNoChange(&raw.DBURL, *dbURLFlag, "db urls")
	if err != nil {
		return err
	}

	err = setStringNoChange(&raw.LogLevel, *logLevelFlag, "log levels")
	if err != nil {
		return err
	}

	err = setStringNoChange(&raw.PublicAPIURL, *publicAPIURLFlag,
		"public api urls")
	if err != nil {
		return err
	}

	err = setStringNoChange(&raw.PublicIDPURL, *publicIDPURLFlag,
		"public idp urls")
	if err != nil {
		return err
	}

	err = setStringSliceNoChange(&raw.ClientHosts, *clientHostsFlag,
		"client hosts")
	if err != nil {
		return err
	}

	return nil
}

// setNoChangeEnvVars will set all of the values provided as environment vars
// as long as they don't change existing non-empty values in Configs
func (raw *rawConfigs) setNoChangeEnvVars() error {

	err := raw.setNoChangeDatabaseEnvVars()
	if err != nil {
		return err
	}

	err = setStringNoChange(&raw.PublicAPIURL, publicAPIURLEnv,
		"public api urls")
	if err != nil {
		return err
	}

	err = setStringNoChange(&raw.PublicIDPURL, publicIDPURLEnv,
		"public idp urls")
	if err != nil {
		return err
	}

	err = setStringSliceNoChange(&raw.ClientHosts, clientHostsEnv,
		"client hosts")
	if err != nil {
		return err
	}

	return nil
}

func (raw *rawConfigs) setNoChangeDatabaseEnvVars() error {
	// if all of the env vars are unset, do nothing
	if dbDriverEnv == "" &&
		dbHostEnv == "" &&
		dbPortEnv == "" &&
		dbSSLEnv == "" &&
		dbUserEnv == "" &&
		dbPassEnv == "" &&
		dbDBNameEnv == "" {
		return nil
	}

	envVarDB := &url.URL{}
	if dbDriverEnv != "" {
		envVarDB.Scheme = dbDriverEnv
	}

	if dbUserEnv != "" {
		envVarDB.User = url.UserPassword(dbUserEnv, dbPassEnv)
	}

	if dbHostEnv != "" {
		envVarDB.Host = dbHostEnv
	}

	if dbPortEnv != "" {
		envVarDB.Host += ":" + dbPortEnv
	}

	if dbDBNameEnv != "" {
		envVarDB.Path = dbDBNameEnv
	}

	if dbSSLEnv != "" {
		envVarDB.RawQuery = "sslmode=" + dbSSLEnv
	}

	return setStringNoChange(&raw.DBURL, envVarDB.String(), "db urls")
}

// setStringNoChange will set the config field to the value (if there is a
// value), and if it doesn't change an existing configuration
func setStringNoChange(configPtr *string, val string, errName string) error {
	if val == "" {
		return nil
	}

	if *configPtr == "" {
		*configPtr = val
		return nil
	}

	if *configPtr == val {
		return nil
	}

	return configErr.New("%s %q and %q don't match", errName, *configPtr, val)
}

// setStringSliceNoChange will set the config field to the csv value (if there
// is a value), and if it doesn't change an existing configuration
func setStringSliceNoChange(configPtr *[]string, val string,
	errName string) error {
	if val == "" {
		return nil
	}

	csv := strings.Split(val, ",")
	if len(*configPtr) == 0 {
		*configPtr = csv
		return nil
	}

	if len(*configPtr) != len(csv) {
		return configErr.New("%s %q and %q don't match", errName, *configPtr, val)
	}

	for i := range *configPtr {
		if (*configPtr)[i] != csv[i] {
			return configErr.New("%s %q and %q don't match", errName, *configPtr, val)
		}
	}

	return nil
}

func (raw *rawConfigs) validate() (*Configs, error) {
	if raw.DBURL == "" {
		return nil, configErr.New("db_url unconfigured")
	}
	if raw.APISlug == "" {
		return nil, configErr.New("api_slug unconfigured")
	}
	if raw.APIAddress == "" {
		return nil, configErr.New("api_addr unconfigured")
	}
	//if raw.MetricAddress == "" {
	//	return nil, configErr.New("metric_addr unconfigured")
	//}
	if raw.GracefulShutdownTimeout == 0 {
		return nil, configErr.New("graceful_shutdown_timeout_sec unconfigured")
	}
	if raw.WriteTimeout == 0 {
		return nil, configErr.New("write_sec unconfigured")
	}
	if raw.ReadTimeout == 0 {
		return nil, configErr.New("read_sec unconfigured")
	}
	if raw.IdleTimeout == 0 {
		return nil, configErr.New("idle_sec unconfigured")
	}
	if raw.IDPPasswordSalt == "" {
		return nil, configErr.New("idp_password_salt unconfigured")
	}
	if raw.IDPClientID == "" {
		return nil, configErr.New("idp_client_id unconfigured")
	}
	if raw.IDPClientSecret == "" {
		return nil, configErr.New("idp_client_secret unconfigured")
	}
	if raw.LogLevel == "" {
		return nil, configErr.New("loglevel unconfigured")
	}
	if len(raw.ClientHosts) == 0 {
		return nil, configErr.New("client_hosts unconfigured")
	}
	if raw.PublicAPIURL == "" {
		return nil, configErr.New("public_api_url unconfigured")
	}
	if raw.PublicIDPURL == "" {
		return nil, configErr.New("public_idp_url unconfigured")
	}

	dbURL, err := url.Parse(raw.DBURL)
	if err != nil {
		return nil, err
	}

	publicAPIURL, err := url.Parse(raw.PublicAPIURL)
	if err != nil {
		return nil, err
	}

	publicIDPURL, err := url.Parse(raw.PublicIDPURL)
	if err != nil {
		return nil, err
	}

	grace := time.Second * time.Duration(raw.GracefulShutdownTimeout)
	write := time.Second * time.Duration(raw.WriteTimeout)
	read := time.Second * time.Duration(raw.ReadTimeout)
	idle := time.Second * time.Duration(raw.IdleTimeout)

	loglevel, err := logrus.ParseLevel(raw.LogLevel)
	if err != nil {
		return nil, err
	}

	clientHosts := make([]*url.URL, 0, len(raw.ClientHosts))
	for _, ch := range raw.ClientHosts {
		clientHost, err := url.Parse(ch)
		if err != nil {
			return nil, err
		}
		clientHosts = append(clientHosts, clientHost)
	}

	return &Configs{
		Version:    "<unset>",
		DBURL:      dbURL,
		APISlug:    raw.APISlug,
		APIAddress: raw.APIAddress,
		IDPAddress: raw.IDPAddress,
		//MetricAddress:           raw.MetricAddress,
		GracefulShutdownTimeout: grace,
		WriteTimeout:            write,
		ReadTimeout:             read,
		IdleTimeout:             idle,
		IDPPasswordSalt:         raw.IDPPasswordSalt,
		IDPClientID:             raw.IDPClientID,
		IDPClientSecret:         raw.IDPClientSecret,
		LogLevel:                loglevel,
		DeveloperMode:           raw.DeveloperMode,
		InsecureRequestsMode:    raw.InsecureRequestsMode,
		ClientHosts:             clientHosts,
		PublicAPIURL:            publicAPIURL,
		PublicIDPURL:            publicIDPURL,
	}, nil
}
