package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"democart/config"
	"democart/database"
	"democart/util"
)

// common tools used when testing the server

type serverTest struct {
	*testing.T
	server *Server
	dbURL  string
}

func newServerTest(t *testing.T) (context.Context, *serverTest) {
	// https://www.sqlite.org/inmemorydb.html
	testDBURL, err := url.Parse("sqlite3::memory:")
	assert.NoError(t, err)
	testDB, err := database.Connect(testDBURL, nil)
	assert.NoError(t, err)
	c := &config.Configs{
		IDPPasswordSalt: "salt",
		IDPClientID:     "idpid",
		IDPClientSecret: "idpsecret",
		DeveloperMode:   false,
		ClientHosts:     nil,
	}
	return context.Background(), &serverTest{
		T:      t,
		server: New(testDB, c),
		dbURL:  testDBURL.String(),
	}
}

func (st *serverTest) addNewSession(ctx context.Context,
	email string) context.Context {
	return SetCtxSession(ctx, newSessionUser(ctx, st, email))
}

func (st *serverTest) cleanup() {
	assert.NoError(st, st.server.DB.Close())
}

// newSessionUser manually creates a user with an active session and returns
// their access token
func newSessionUser(ctx context.Context, st *serverTest,
	email string) *database.Session {
	user, err := st.server.DB.Create_User(ctx,
		database.User_Id(util.MustUUID4()),
		database.User_Email(email),
		database.User_ProfileUrl(""),
		database.User_FullName(""))
	assert.NoError(st, err)

	session, err := st.server.DB.Create_Session(ctx,
		database.Session_Id(util.MustUUID4()),
		database.Session_IdToken(util.MustUUID4()),
		database.Session_AccessToken(util.MustUUID4()),
		database.Session_RefreshToken(util.MustUUID4()),
		database.Session_AccessTokenExpiry(util.UTCNow().Add(time.Minute)),
		database.Session_DeviceName("unittest"),
		database.Session_Create_Fields{
			UserPk: database.Session_UserPk(user.Pk),
		})
	assert.NoError(st, err)
	return session
}

// newItem manually creates an item
func newItem(ctx context.Context, st *serverTest,
	description string, rq int) *database.Item {
	item, err := st.server.DB.Create_Item(ctx,
		database.Item_Id(util.MustUUID4()),
		database.Item_Price(10),
		database.Item_Description(description),
		database.Item_ImageUrl(""),
		database.Item_RemainingQuantity(rq),
		database.Item_Create_Fields{})
	assert.NoError(st, err)
	return item
}

func jsonPostRequest(st *serverTest, target string,
	body interface{}) *http.Request {
	buf, err := json.Marshal(body)
	assert.NoError(st, err)
	r := httptest.NewRequest(http.MethodPost, target, bytes.NewReader(buf))
	r.Header.Set("Content-Type", "application/json")
	return r
}
