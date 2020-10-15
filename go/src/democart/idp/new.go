package idp

import (
	"crypto/sha256"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"democart/config"
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

func (i *IDP) Close() error {
	return i.DB.Close()
}

func router(i *IDP) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	mw := h.MiddlewareChain()

	r.Method("GET", "/idplogin", mw.Bytes(i.Login))
	r.Method("POST", "/idplogincomplete", mw.JSON(i.LoginComplete))

	r.Method("GET", "/idpsignup", mw.Bytes(i.Signup))
	r.Method("POST", "/idpsignupcomplete", mw.JSON(i.SignupComplete))

	r.Method("POST", "/idptoken", mw.JSON(i.TokenExchange))
	return r
}

// NewHTTPServer constructs a new http.Server to listen for connections and
// serve responses as defined by the IDP's ServeHTTP defined above.
func NewHTTPServer(configs *config.Configs) (*IDP, *http.Server, error) {

	// just connect to the existing db instead of trying to separate it out like
	// we would want to do irl
	// TODO(sam): pass through database configs
	db, err := database.Connect(configs.DBURL, nil)
	if err != nil {
		return nil, nil, err
	}

	idpClient := New(configs.IDPPasswordSalt, db)
	return idpClient, &http.Server{
		Addr:         configs.IDPAddress,
		WriteTimeout: configs.WriteTimeout,
		ReadTimeout:  configs.ReadTimeout,
		IdleTimeout:  configs.IdleTimeout,
		Handler:      idpClient,
	}, nil
}
