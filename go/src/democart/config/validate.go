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
	dbURLFlag       = flag.String("db_url", "", "database url") // flag override
	logLevelFlag    = flag.String("loglevel", "", "log level")  // flag override

	// env var overrides
	dbDriverEnv = os.Getenv("DATABASE_DRIVER")
	dbHostEnv   = os.Getenv("DATABASE_HOST")
	dbPortEnv   = os.Getenv("DATABASE_PORT")
	dbSSLEnv    = os.Getenv("DATABASE_SSLMODE")
	dbUserEnv   = os.Getenv("DATABASE_USER")
	dbPassEnv   = os.Getenv("DATABASE_PASSWORD")
	dbDBNameEnv = os.Getenv("DATABASE_NAME")

	configErr = errs.Class("configuration")
)

type Configs struct {
	Version    string
	DBURL      *url.URL
	APISlug    string // shortname for the api server
	APIAddress string // the port the api server is listening on localhost
	//IDPAddress           string // the port the idp server is listening on
	//ExposedIDPURL        string // the scheme and host where the idp is publicly accessible
	//MetricAddress           string // the port the metric server is listening on
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
	ExposedURL              *url.URL // the scheme and host where this api is publicily accessible
}

func (c *Configs) SetVersion(v string) { c.Version = v }

// Parse will set the configuration values pulled from the provided config
// file, any config flags, and any environment variables. env vars have highest
// priority, then config flags, then the config file. If any values overlap and
// *DON'T* match, an error is thrown.
func Parse() (*Configs, error) {
	flag.Parse()

	raw := rawConfigs{}
	err := raw.setConfigFiles()
	if err != nil {
		return nil, err
	}

	err = raw.setNoChangeConfigFlags()
	if err != nil {
		return nil, err
	}

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
	ExposedURL              string   `hcl:"exposed_url"`
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
	if dbURLFlag != nil && *dbURLFlag != "" {
		if raw.DBURL == "" {
			raw.DBURL = *dbURLFlag
		} else if raw.DBURL != *dbURLFlag {
			return configErr.New("db urls %q and %q don't match", raw.DBURL,
				*dbURLFlag)
		}
	}

	if logLevelFlag != nil && *logLevelFlag != "" {
		if raw.LogLevel == "" {
			raw.LogLevel = *logLevelFlag
		} else if raw.LogLevel != *logLevelFlag {
			return configErr.New("log level %q and %q don't match", raw.LogLevel,
				*logLevelFlag)
		}
	}

	return nil
}

// setNoChangeEnvVars will set all of the values provided as environment vars
// as long as they don't change existing non-empty values in Configs
func (raw *rawConfigs) setNoChangeEnvVars() error {

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

	envVarDBString := envVarDB.String()
	if raw.DBURL == "" {
		raw.DBURL = envVarDBString
	} else if raw.DBURL != envVarDBString {
		return configErr.New("db urls %q and env %q don't match", raw.DBURL,
			envVarDBString)
	}

	return nil
}

func (raw *rawConfigs) validate() (*Configs, error) {
	if raw.DBURL == "" {
		return nil, configErr.New("db_url misconfigured")
	}
	if raw.APISlug == "" {
		return nil, configErr.New("api_slug misconfigured")
	}
	if raw.APIAddress == "" {
		return nil, configErr.New("api_addr misconfigured")
	}
	//if raw.MetricAddress == "" {
	//	return nil, configErr.New("metric_addr misconfigured")
	//}
	if raw.GracefulShutdownTimeout == 0 {
		return nil, configErr.New("graceful_shutdown_timeout_sec misconfigured")
	}
	if raw.WriteTimeout == 0 {
		return nil, configErr.New("write_sec misconfigured")
	}
	if raw.ReadTimeout == 0 {
		return nil, configErr.New("read_sec misconfigured")
	}
	if raw.IdleTimeout == 0 {
		return nil, configErr.New("idle_sec misconfigured")
	}
	if raw.IDPPasswordSalt == "" {
		return nil, configErr.New("idp_password_salt misconfigured")
	}
	if raw.IDPClientID == "" {
		return nil, configErr.New("idp_client_id misconfigured")
	}
	if raw.IDPClientSecret == "" {
		return nil, configErr.New("idp_client_secret misconfigured")
	}
	if raw.LogLevel == "" {
		return nil, configErr.New("loglevel misconfigured")
	}
	if raw.ExposedURL == "" {
		return nil, configErr.New("exposed_url misconfigured")
	}

	dbURL, err := url.Parse(raw.DBURL)
	if err != nil {
		return nil, err
	}

	exposedURL, err := url.Parse(raw.ExposedURL)
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

	if len(raw.ClientHosts) == 0 {
		return nil, configErr.New("client_hosts misconfigured")
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
		ExposedURL:              exposedURL,
	}, nil
}
