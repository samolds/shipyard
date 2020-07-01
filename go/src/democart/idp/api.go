package idp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"democart/database"
	he "democart/httperror"
	"democart/util"
)

var (
	validCodeDuration          = -30 * time.Minute
	defaultTokenExpiryDuration = 24 * time.Hour
)

// TODO(sam): doesn't protect against CSRF attacks. again, this is not a real
// auth system.
const formFmt = `<h1>%s</h1>
<form method="post" action="%s">
	<label for="email">Email</label>
	<input type="email" id="email" name="email">
	<label for="password">Password</label>
	<input type="password" id="password" name="password">
	<button type="submit">%s</button>
</form>`

func emailPasswordForm(title, action string, q url.Values) []byte {
	if q != nil {
		q.Del("err") // make sure there isn't some lingering err somehow
		action += "?" + q.Encode()
	}
	return []byte(fmt.Sprintf(formFmt, title, action, title))
}

// Login returns HTML to the user to provide them with a way to log in with
// their email and password
func (i *IDP) Login(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	q := r.URL.Query()
	form := emailPasswordForm("Login", "/fakeidp/idplogincomplete", q)
	return form, nil
}

// LoginComplete accepts the form provided email and password, and makes sure
// that the email/password_hash pair exists before redirecting back to the
// redirect_uri provided at the beginning of the login flow with a code.
func (i *IDP) LoginComplete(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	find := func(ctx context.Context, email string, pwdHash []byte) (string,
		error) {
		ep, err := i.DB.Find_EmailPassword_By_Email_And_PasswordHash(ctx,
			database.EmailPassword_Email(email),
			database.EmailPassword_PasswordHash(pwdHash))
		if err != nil {
			return "", err
		}

		if ep == nil {
			return "", he.NotFound.New("that is not a valid email/password combo")
		}

		newCode := util.MustUUID4()
		err = i.DB.UpdateNoReturn_EmailPassword_By_Pk(ctx,
			database.EmailPassword_Pk(ep.Pk),
			database.EmailPassword_Update_Fields{
				Code: database.EmailPassword_Code(newCode),
			})
		if err != nil {
			return "", err
		}
		return newCode, nil
	}

	return i.complete(ctx, w, r, find)
}

// Signup returns HTML to the user to provide them with a way to sign up with
// their email and password
func (i *IDP) Signup(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	q := r.URL.Query()
	form := emailPasswordForm("Signup", "/fakeidp/idpsignupcomplete", q)
	return form, nil
}

// SignupComplete accepts the form provided email and password, and creates
// a unique email/password_hash pair in the db before redirect back to the
// redirect_uri provided at the beginning of the signup flow with a code.
func (i *IDP) SignupComplete(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	create := func(ctx context.Context, email string, pwdHash []byte) (string,
		error) {
		code := util.MustUUID4()
		err := i.DB.CreateNoReturn_EmailPassword(ctx,
			database.EmailPassword_Email(email),
			database.EmailPassword_PasswordHash(pwdHash),
			database.EmailPassword_Code(code))
		if err != nil {
			return "", he.BadRequest.Wrap(err) // expected error is duplicate email
		}
		return code, nil
	}
	return i.complete(ctx, w, r, create)
}

type codeGetter func(context.Context, string, []byte) (string, error)

func (i *IDP) complete(ctx context.Context, w http.ResponseWriter,
	r *http.Request, getCode codeGetter) (interface{}, error) {

	err := r.ParseForm()
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}
	email := r.FormValue("email")
	password := r.FormValue("password")
	pwdHash := i.passwordHash(password)

	// TODO(sam): why is a whole bunch of extra stuff getting nested here. is
	// there a global value getting mutated??
	redirectURIRaw, err := url.Parse(r.URL.Query().Get("redirect_uri"))
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}

	// allows the client app to pass additional query params in the redirect_uri
	// that will be included with the code and state
	q := redirectURIRaw.Query()
	q.Del("err") // make sure there isn't some lingering err somehow
	redirectURIRaw.RawQuery = ""
	redirectURIRaw.Fragment = ""
	redirectURI := redirectURIRaw.String() + "?"

	code, err := getCode(ctx, email, pwdHash)
	if err != nil {
		logrus.Debugf("idp completion error: %s. redirecting...", err)
		q.Set("err", fmt.Sprintf("%s", err))
		http.Redirect(w, r, redirectURI+q.Encode(), http.StatusFound)
		return nil, err
	}

	// build the redirect with the code
	// TODO(sam): include the redirect back to the backend server
	q.Set("code", code)
	q.Set("state", r.URL.Query().Get("state"))

	http.Redirect(w, r, redirectURI+q.Encode(), http.StatusFound)
	return nil, nil
}

type tokenRequestBody struct {
	Code         string `json:"code,omitempty"`
	ClientID     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
	GrantType    string `json:"grant_type,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	RedirectURI  string `json:"redirect_uri,omitempty"`
}

type tokenData struct {
	Email       string       `json:"email"`
	IDToken     string       `json:"id_token"`
	Oauth2Token oauth2.Token `json:"oauth2_token"`
}

// TokenExchange accepts a POST body with a code or a refresh token and
// responds with an unexpired Oauth2 Token
func (i *IDP) TokenExchange(ctx context.Context, w http.ResponseWriter,
	r *http.Request) (interface{}, error) {

	// TODO(sam): this should also check client id/secret and stuff. but since
	// this is a fake id, just naively trust the provided code.
	var token tokenRequestBody
	err := json.NewDecoder(r.Body).Decode(&token)
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}

	lastLoginMoreRecentThan := util.UTCNow().Add(validCodeDuration)
	ep, err := i.DB.Find_EmailPassword_By_Code_And_LastLogin_Greater(ctx,
		database.EmailPassword_Code(token.Code),
		database.EmailPassword_LastLogin(lastLoginMoreRecentThan))
	if err != nil {
		return nil, he.Unexpected.Wrap(err)
	}

	if ep == nil {
		return nil, he.NotFound.New("invalid code")
	}

	// the code is one time use
	err = i.DB.UpdateNoReturn_EmailPassword_By_Pk(ctx,
		database.EmailPassword_Pk(ep.Pk),
		database.EmailPassword_Update_Fields{Code: database.EmailPassword_Code("")})
	if err != nil {
		return nil, err
	}

	// TODO(sam): this is a bit funky. it would be better to use JWTs
	t := tokenData{
		Email:   ep.Email,
		IDToken: util.MustUUID4(),
		Oauth2Token: oauth2.Token{
			AccessToken:  util.MustUUID4(),
			TokenType:    "bearer",
			RefreshToken: util.MustUUID4(),
			Expiry:       util.UTCNow().Add(defaultTokenExpiryDuration),
		},
	}

	return t, nil
}
