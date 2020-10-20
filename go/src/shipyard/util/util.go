package util

import (
	"time"

	"github.com/google/uuid"
	"github.com/zeebo/errs"
)

func UTCNow() time.Time {
	return time.Now().UTC()
}

func MustUUID4() string {
	// NewRandom returns a Version 4 UUID
	u, err := uuid.NewRandom()
	if err != nil {
		panic(errs.Wrap(err))
	}
	return u.String()
}
