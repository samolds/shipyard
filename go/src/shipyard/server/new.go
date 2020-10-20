package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/sirupsen/logrus"

	"shipyard/config"
	"shipyard/database"
	h "shipyard/handler"
)

type Server struct {
	DB     *database.DB
	Config *config.Configs
	log    *logrus.Entry
	router http.Handler
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func New(db *database.DB, configs *config.Configs) *Server {
	s := &Server{
		DB:     db,
		Config: configs,
		log:    logrus.WithField("version", configs.Version),
	}
	s.router = router(s)
	return s
}

func (s *Server) Close() error {
	return s.DB.Close()
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

	return r
}

// NewHTTPServer constructs a new http.Server to listen for connections and
// serve responses as defined by the Server's ServeHTTP defined above.
func NewHTTPServer(configs *config.Configs,
	metricMiddleware h.MiddlewareWrapper) (*Server, *http.Server, error) {

	// TODO(sam): pass through database configs
	db, err := database.Connect(configs.DBURL, nil)
	if err != nil {
		return nil, nil, err
	}

	apiClient := New(db, configs)

	var apiHandler http.Handler = apiClient
	if metricMiddleware != nil {
		apiHandler = metricMiddleware(apiHandler)
	}

	return apiClient, &http.Server{
		Addr:         configs.APIAddress,
		WriteTimeout: configs.WriteTimeout,
		ReadTimeout:  configs.ReadTimeout,
		IdleTimeout:  configs.IdleTimeout,
		Handler:      apiHandler,
	}, nil
}
