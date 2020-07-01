package idp

import (
	"crypto/sha256"
	"net/http"

	"github.com/go-chi/chi"

	"democart/database"
	h "democart/handler"
)

type IDP struct {
	salt   string
	DB     *database.DB
	router http.Handler
}

func (i *IDP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	i.router.ServeHTTP(w, r)
}

func (i *IDP) passwordHash(password string) []byte {
	h := sha256.New()
	h.Write([]byte(i.salt + password))
	return h.Sum(nil)
}

func New(salt string, db *database.DB) *IDP {
	i := &IDP{salt: salt, DB: db}
	i.router = router(i)
	return i
}

func router(i *IDP) http.Handler {
	r := chi.NewRouter()

	mw := h.MiddlewareChain()

	r.Method("GET", "/idplogin", mw.Bytes(i.Login))
	r.Method("POST", "/idplogincomplete", mw.JSON(i.LoginComplete))

	r.Method("GET", "/idpsignup", mw.Bytes(i.Signup))
	r.Method("POST", "/idpsignupcomplete", mw.JSON(i.SignupComplete))

	r.Method("POST", "/idptoken", mw.JSON(i.TokenExchange))
	return r
}
