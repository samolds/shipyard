package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"
	"golang.org/x/oauth2"

	"democart/database"
	he "democart/httperror"
	"democart/util"
)

func (s *Server) idpURL(next string) string {
	// TODO(sam): this should be composed from the request url, not the config
	// server url. this is because the exposed route (in docker) might be
	// different than what url is provided in the config
	root, _ := url.Parse(s.Config.ServerURL.String())
	root.Path = path.Join(root.Path, fakeIDPPath, next)
	return root.String()
}

func (s *Server) isWhitelistedClientApp(referrer string) (bool, *url.URL) {
	referringURL, err := url.Parse(referrer)
	if err != nil {
		logrus.Debugf("%q - failed to parse: %s", referrer, err)
		return false, nil
	}

	hostStr := referringURL.Host
	for _, h := range s.Config.ClientHosts {
		if hostStr == h.Host {
			return true, referringURL
		}
	}

	logrus.Debugf("host %q not found in whitelist %+v", hostStr,
		s.Config.ClientHosts)
	return false, nil
}

// Signup will redirect the requester to the configured identity provider where
// the user can provide their credentials and receive a code. This code needs
// to be provided back to this resource server for a valid access token
func (s *Server) Login(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	// build the redirect_uri that the code needs to be given to after a
	// successful credential exchange
	return s.beginAuth(ctx, w, r, s.idpURL("/idplogin"), "/auth/logincomplete")
}

func (s *Server) LoginComplete(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	findUser := func(ctx context.Context, email string) (*database.User, error) {
		user, err := s.DB.Find_User_By_Email(ctx, database.User_Email(email))
		if err != nil {
			return nil, errs.Wrap(err)
		}

		if user == nil {
			return nil, he.NotFound.New("%q doesn't exist. please sign up", email)
		}
		return user, nil
	}

	return s.completeAuth(ctx, w, r, findUser)
}

// Signup will redirect the requester to the configured identity provider where
// the user can provide their credentials and receive a code. This code needs
// to be provided back to this resource server for a valid access token
func (s *Server) Signup(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	// build the redirect_uri that the code needs to be given to after a
	// successful credential exchange
	return s.beginAuth(ctx, w, r, s.idpURL("/idpsignup"), "/auth/signupcomplete")
}

func (s *Server) SignupComplete(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	makeUser := func(ctx context.Context, email string) (*database.User, error) {
		u, err := s.DB.Create_User(ctx, database.User_Id(util.MustUUID4()),
			database.User_Email(email), database.User_ProfileUrl(""),
			database.User_FullName(""))
		if err != nil {
			return nil, errs.Wrap(err)
		}
		return u, nil
	}

	return s.completeAuth(ctx, w, r, makeUser)
}

// beginAuth will redirect to the identity provider with a redirect_uri
// provided to send the user back to their own app that made the initial
// request, with a code and an additional redirect_uri to exchange the code
// with the resource server
func (s *Server) beginAuth(ctx context.Context, w http.ResponseWriter,
	r *http.Request, idpURI, finalRedirectURI string) (interface{}, error) {

	// TODO(sam): this should be composed from the request url, not the config
	// server url. this is because the exposed route (in docker) might be
	// different than what url is provided in the config
	codeExchanger, err := url.Parse(s.Config.ServerURL.String())
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}

	codeExchanger.Path = path.Join(finalRedirectURI)
	codeExchange := codeExchanger.String()
	referrer := r.Header.Get("referer")
	if ok, referringURL := s.isWhitelistedClientApp(referrer); ok {
		// if the referrer is an approved client app, build up the redirect_uri to
		// redirect back to the client app, with an additional redirect_uri back to
		// the code exchanging auth completion endpoint
		done := url.Values{}
		done.Set("redirect_uri", codeExchange)
		referringURL.RawQuery = done.Encode()
		referringURL.Fragment = ""
		codeExchange = referringURL.String()
	}

	q := url.Values{}
	q.Set("response_type", "code")
	q.Set("client_id", s.Config.IDPClientID)
	q.Set("redirect_uri", codeExchange)
	q.Set("scope", "fake_scope")
	q.Set("state", util.MustUUID4())

	http.Redirect(w, r, idpURI+"?"+q.Encode(), http.StatusFound)
	return nil, nil
}

type tokenPostBody struct {
	Code         string `json:"code,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	GrantType    string `json:"grant_type,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
}

// TODO(sam): this could be done better by using JWTs or something
type tokenReceiveBody struct {
	Email       string       `json:"email"`
	IDToken     string       `json:"id_token"`
	Oauth2Token oauth2.Token `json:"oauth2_token"`
}

type userGetter func(context.Context, string) (*database.User, error)

func (s *Server) completeAuth(ctx context.Context, w http.ResponseWriter,
	r *http.Request, getUser userGetter) (interface{}, error) {

	idpErr := r.URL.Query().Get("err")
	if idpErr != "" {
		return nil, he.Unexpected.New(idpErr)
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		return nil, he.Unexpected.New("unexpected code during login: %q", code)
	}

	state := r.URL.Query().Get("state")
	if state == "" {
		return nil, he.Unexpected.New("unexpected state during login: %q", state)
	}

	exchange := tokenPostBody{
		Code:      code,
		GrantType: "authorization_code",
	}

	buf := new(bytes.Buffer)
	err := json.NewEncoder(buf).Encode(&exchange)
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}

	// TODO(sam): don't use the default client. wrap so that testing isn't
	// actually making http requests
	resp, err := http.Post(s.idpURL("/idptoken"), "application/json", buf)
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, he.Unexpected.New("unexpected status: " + resp.Status)
	}

	var token tokenReceiveBody
	err = json.NewDecoder(resp.Body).Decode(&token)
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}

	deviceName := r.Header.Get("user-agent")
	user, err := getUser(ctx, token.Email)
	if err != nil {
		return nil, err
	}

	session, err := s.DB.Create_Session(ctx,
		database.Session_Id(util.MustUUID4()),
		database.Session_IdToken(token.IDToken),
		database.Session_AccessToken(token.Oauth2Token.AccessToken),
		database.Session_RefreshToken(token.Oauth2Token.RefreshToken),
		database.Session_AccessTokenExpiry(token.Oauth2Token.Expiry),
		database.Session_DeviceName(deviceName),
		database.Session_Create_Fields{
			UserPk: database.Session_UserPk(user.Pk),
		})
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}

	jsonResp := &RootJSON{
		User:     apiUser(user),
		Session:  apiSession(session),
		Response: "successfully signed in",
	}
	return jsonResp, nil
}

// Logout will delete the current session
func (s *Server) Logout(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	ss, err := GetCtxSession(ctx)
	if err != nil {
		return nil, err
	}

	_, err = s.DB.Delete_Session_By_Pk(ctx, database.Session_Pk(ss.Pk))
	if err != nil {
		return nil, err
	}

	jsonResp := &RootJSON{
		Response: "successfully logged out",
	}
	return jsonResp, nil
}
