package server

import (
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"

	"democart/database"
	h "democart/handler"
	"democart/idp"
)

var (
	fakeIDPPath = "/fakeidp"
)

type Config struct {
	IDPPasswordSalt string
	IDPClientID     string
	IDPClientSecret string
	ServerURL       *url.URL
	DeveloperMode   bool
	ClientHosts     []*url.URL
}

type Server struct {
	DB     *database.DB
	router http.Handler
	Config *Config
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func New(db *database.DB, config *Config) *Server {
	s := &Server{DB: db, Config: config}
	s.router = router(s)
	return s
}

func router(s *Server) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	clientHosts := make([]string, 0, len(s.Config.ClientHosts))
	for _, h := range s.Config.ClientHosts {
		clientHosts = append(clientHosts, h.String())
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: clientHosts,
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type",
			"X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any major browsers
	}))

	mw := h.MiddlewareChain()
	r.Method("GET", "/", mw.JSON(s.Health))

	// all authentication routes must be performed unauthenticated
	authRoutes := chi.NewRouter()
	authMW := mw.Append(s.Unauthenticated) // add middleware
	authRoutes.Method("GET", "/signup", authMW.JSON(s.Signup))
	authRoutes.Method("GET", "/signupcomplete", authMW.JSON(s.SignupComplete))
	authRoutes.Method("GET", "/login", authMW.JSON(s.Login))
	authRoutes.Method("GET", "/logincomplete", authMW.JSON(s.LoginComplete))
	authRoutes.Method("GET", "/logout", mw.Append(s.Authenticated).JSON(s.Logout))
	r.Mount("/auth", authRoutes)

	// all api routes must be performed authenticated
	apiRoutes := chi.NewRouter()
	apiMW := mw.Append(s.Authenticated) // add middleware
	apiRoutes.Method("GET", "/", apiMW.JSON(s.UserProfile))
	apiRoutes.Method("POST", "/address", apiMW.JSON(s.AddAddress))
	apiRoutes.Method("GET", "/item", mw.JSON(s.ListItem)) // no auth
	apiRoutes.Method("POST", "/item", apiMW.JSON(s.AddItem))
	apiRoutes.Method("POST", "/item/{itemID}", apiMW.JSON(s.UpdateItem))
	apiRoutes.Method("GET", "/cart", apiMW.JSON(s.ListCart))
	apiRoutes.Method("POST", "/cart", apiMW.JSON(s.AddCart))
	apiRoutes.Method("POST", "/cart/{cartItemID}", apiMW.JSON(s.UpdateCart))
	apiRoutes.Method("GET", "/order", apiMW.JSON(s.ListOrder))
	apiRoutes.Method("POST", "/order", apiMW.JSON(s.AddOrder))
	r.Mount("/api", apiRoutes)

	// this is a fake identity provider that should be easily swappable for
	// something real, like Auth0. All that's needed is to redirect to a proper
	// idp instead of this one, and make sure to include the necessary client
	// ids/secrets/etc.
	idpRoutes := idp.New(s.Config.IDPPasswordSalt, s.DB)
	r.Mount(fakeIDPPath, idpRoutes)

	return r
}
