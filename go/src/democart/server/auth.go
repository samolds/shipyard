package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"

	"democart/database"
	"democart/handler"
	he "democart/httperror"
	"democart/util"
)

const (
	sessionKey = iota
)

type sessionCtxKey int

func SetCtxSession(ctx context.Context,
	session *database.Session) context.Context {
	return context.WithValue(ctx, sessionCtxKey(sessionKey), session)
}

func GetCtxSession(ctx context.Context) (*database.Session, error) {
	ss, ok := ctx.Value(sessionCtxKey(sessionKey)).(*database.Session)
	if !ok || ss == nil {
		return nil, he.Unauthenticated.New("session not set in context")
	}
	return ss, nil
}

func (s *Server) Unauthenticated(h handler.Handler) handler.Handler {
	return handler.Handler(func(ctx context.Context, w http.ResponseWriter,
		r *http.Request) (interface{}, error) {

		logrus.Debugf("checking that the request is not authenticated")

		authorizationHeader := r.Header.Get("authorization")
		if authorizationHeader == "" {
			// opposite logic. no header is what we're looking for
			logrus.Debugf("authorization header not found. good")
			return h(ctx, w, r)
		}

		parts := strings.Fields(authorizationHeader)
		if len(parts) != 2 {
			// opposite logic. unexpected header format
			logrus.Debugf("authorization header has no 'Bearer <token>'. good")
			return h(ctx, w, r)
		}

		token := parts[1]
		logrus.Debugf("found token %q", token)

		ss, err := s.DB.Find_Session_By_AccessToken(ctx,
			database.Session_AccessToken(token))
		if err != nil {
			return nil, err
		}

		if ss == nil {
			// opposite logic. no valid session is found
			logrus.Debugf("no active session found for token %q. good", token)
			return h(ctx, w, r)
		}

		if util.UTCNow().After(ss.AccessTokenExpiry) {
			// TODO(sam): delete this session from the db
			// opposite logic. no valid session is found
			logrus.Debugf("%q is an expired token. good", token)
			return h(ctx, w, r)
		}

		logrus.Debugf("%q is an active token", token)

		return nil, he.BadRequest.New("already logged in")
	})
}

func (s *Server) Authenticated(h handler.Handler) handler.Handler {
	return handler.Handler(func(ctx context.Context, w http.ResponseWriter,
		r *http.Request) (interface{}, error) {

		logrus.Debugf("checking that the request is authenticated")

		authorizationHeader := r.Header.Get("authorization")
		if authorizationHeader == "" {
			return nil, he.Unauthenticated.New("no authorization header")
		}

		parts := strings.Fields(authorizationHeader)
		if len(parts) != 2 {
			return nil, he.Unauthenticated.New("bad authorization header")
		}

		token := parts[1]

		ss, err := s.DB.Find_Session_By_AccessToken(ctx,
			database.Session_AccessToken(token))
		if err != nil {
			return nil, he.Unexpected.Wrap(err)
		}

		if ss == nil {
			return nil, he.Unauthenticated.New("expired session. please login")
		}

		if util.UTCNow().After(ss.AccessTokenExpiry) {
			// TODO(sam): delete this session from the db
			return nil, he.Unauthenticated.New("expired session. please login")
		}

		ctx = SetCtxSession(ctx, ss)
		return h(ctx, w, r)
	})
}
