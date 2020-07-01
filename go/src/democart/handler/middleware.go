package handler

// TODO(sam): figure out how to get the middleware to work as http.Handler and
// http.HandlerFunc so that the JSON response type can be agnostic from the
// middleware

// TODO(sam): add an "Exclude" that can be used and chained together right
// before "Bytes" or "JSON" to prevent middleware from being called in a 1-off
// type scenario. instead of allowing for variadic HandlerFunc params, do
// (key,HandlerFunc) pairs where the key is used during the exclude.

type middlewareChain []HandlerFunc

func MiddlewareChain(middlewares ...HandlerFunc) middlewareChain {
	return middlewares
}

func (mwc middlewareChain) Append(middlewares ...HandlerFunc) middlewareChain {
	return append(mwc, middlewares...)
}

func (mwc middlewareChain) Bytes(h Handler) Handler {
	for i := range mwc {
		h = mwc[len(mwc)-1-i](h)
	}
	return h
}

func (mwc middlewareChain) JSON(h Handler) JSON {
	return JSON(mwc.Bytes(h))
}
