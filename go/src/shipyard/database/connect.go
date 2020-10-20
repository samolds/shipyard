package database

import (
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"

	monitor "shipyard/prometheus"
)

const (
	SqliteDriver   = "sqlite3"
	PostgresDriver = "postgres"
)

var (
	dbErr = errs.Class("database")
)

type Config struct {
	MaxOpenConns *int
	MaxIdleConns *int
}

// TODO(sam): this database package needs a lot of love. there should be a
// database interface to make supporting multiple database drivers easier and
// cleaner. all of the switches are gross.
// TODO(sam): add support for migrations! Critical importance.
func Connect(databaseURL *url.URL, c *Config) (*DB, error) {
	// WrapErr is a dbx specific error wrapping hook
	WrapErr = StacktraceWrapAnyError

	// Logger is a dbx specific logging hook. it's called by every dbx query
	Logger = func(format string, args ...interface{}) {
		logrus.Debugf(format, args...)

		go func() {
			// it is perhaps an abuse of this logger to collect metrics with it
			monitor.DatabaseQueryCounter.Inc()
		}()
	}

	// copy the databaseURL
	loggableURL := *databaseURL
	loggableURL.User = nil // don't log username/password
	logrus.Infof("connecting to db: %s", loggableURL.String())

	dbURL := *databaseURL
	driver := strings.ToLower(dbURL.Scheme)
	if driver == "sqlite3" {
		dbURL.Scheme = "file"
	}

	db, err := Open(driver, dbURL.String())
	if err != nil {
		return nil, err
	}

	if isBrandNewDB(db) {
		logrus.Debugf("%s will be initialized", loggableURL.String())
		err = initializeNewDB(driver, db)
		if err != nil {
			return nil, err
		}
	} else {
		logrus.Debugf("%s already exists", loggableURL.String())
	}

	configureDB(db, c)

	// TODO(sam): add support for migrations!
	// err = migrateDB(driver, db)
	// if err != nil {
	//   return nil, err
	// }

	logrus.Infof("connected to database")
	return db, nil
}

func isBrandNewDB(db *DB) bool {
	brandNew := false

	// use the users table as a sentinel of existence
	_, err := db.DB.Exec("SELECT * FROM users LIMIT 1")
	if err != nil {
		brandNew = true
	}

	return brandNew
}

func initializeNewDB(driver string, db *DB) error {
	_, err := db.Exec(db.Schema())
	if err != nil {
		return dbErr.Wrap(err)
	}

	switch driver {
	case PostgresDriver:
	case SqliteDriver:
		_, err = db.DB.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			return dbErr.Wrap(err)
		}
	default:
		return dbErr.New("unexpected driver")
	}

	return nil
}

func configureDB(db *DB, c *Config) {
	if c == nil {
		return
	}

	if c.MaxOpenConns != nil {
		db.DB.SetMaxOpenConns(*c.MaxOpenConns)
	}
	if c.MaxIdleConns != nil {
		db.DB.SetMaxIdleConns(*c.MaxIdleConns)
	}
}
