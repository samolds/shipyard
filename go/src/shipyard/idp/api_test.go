package idp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"shipyard/database"
	he "shipyard/httperror"
)

func TestUnknownUser(baseTest *testing.T) {
	ctx, t := newIDPTest(baseTest)
	defer t.cleanup()

	w := httptest.NewRecorder()
	r := formRequest("email=user&password=password")

	_, err := t.idp.LoginComplete(ctx, w, r)
	assert.Error(t, err)
	assert.True(t, he.NotFound.Has(err))
}

func TestNoDuplicates(baseTest *testing.T) {
	ctx, t := newIDPTest(baseTest)
	defer t.cleanup()

	w := httptest.NewRecorder()
	r := formRequest("email=user&password=password")

	_, err := t.idp.SignupComplete(ctx, w, r)
	assert.NoError(t, err)

	_, err = t.idp.SignupComplete(ctx, w, r)
	assert.Error(t, err)
	assert.True(t, he.BadRequest.Has(err))
}

func TestNewUser(baseTest *testing.T) {
	ctx, t := newIDPTest(baseTest)
	defer t.cleanup()

	w := httptest.NewRecorder()
	r := formRequest("email=user&password=password")

	_, err := t.idp.SignupComplete(ctx, w, r)
	assert.NoError(t, err)

	_, err = t.idp.LoginComplete(ctx, w, r)
	assert.NoError(t, err)
}

///////////////////////////////////////////////////////////////////////////////
// test helpers
///////////////////////////////////////////////////////////////////////////////

func formRequest(body string) *http.Request {
	// the target is not necessary because this request isn't routed, but passed
	// directly to the handler
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

type idpTest struct {
	*testing.T
	idp   *IDP
	dbURL string
}

func newIDPTest(t *testing.T) (context.Context, *idpTest) {
	// https://www.sqlite.org/inmemorydb.html
	testDBURL, err := url.Parse("sqlite3::memory:")
	assert.NoError(t, err)
	testDB, err := database.Connect(testDBURL, nil)
	assert.NoError(t, err)
	return context.Background(), &idpTest{
		T:     t,
		idp:   New("salt", testDB),
		dbURL: testDBURL.String(),
	}
}

func (idpT *idpTest) cleanup() {
	assert.NoError(idpT, idpT.idp.DB.Close())
}
