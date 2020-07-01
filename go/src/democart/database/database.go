package database

import (
	"context"

	"github.com/sirupsen/logrus"
)

// StacktraceWrapAnyError is used by the dbx WrapErr hook to provide stack
// traces on any db error
//
// TODO(sam): more graceful/user-friendly errors should be returned when there
// are unique-key constraint violations, no row found, etc
func StacktraceWrapAnyError(err *Error) error {
	if err == nil {
		return nil
	}

	logrus.WithError(err).Warning("database connection error")
	return dbErr.Wrap(err)
}

// WithTx provides a way to transactionally execute multiple db statements
func (db *DB) WithTx(ctx context.Context,
	fn func(context.Context, *Tx) error) (err error) {

	tx, err := db.Open(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = tx.Commit()
		} else {
			txerr := tx.Rollback() // careful not to shadow "err"
			if txerr != nil {
				logrus.Warningf("transaction rollback error: %s", txerr)
			}
		}
	}()
	return fn(ctx, tx)
}
