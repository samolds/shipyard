package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/zeebo/errs"

	he "shipyard/httperror"
)

func jsonResponse(w http.ResponseWriter, obj interface{}, err error) {
	writeJSONError := func(jsonErr error) {
		statusCode := he.StatusCodeByError(jsonErr)
		w.WriteHeader(statusCode)

		errStr := fmt.Sprintf("%s", jsonErr)
		jsonErrObj := map[string]string{"error": errStr}
		jsonObj, err := json.Marshal(jsonErrObj)
		if err != nil {
			logrus.Warningf("failed to convert %v to json", jsonErrObj)
			w.Header().Set("Content-Type", "text/plain")
			w.Write([]byte(errStr))
			return
		}

		w.Write(jsonObj)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err != nil {
		writeJSONError(err)
		return
	}

	if obj == nil {
		obj = map[string]string{"response": "okay!"}
	}

	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		writeJSONError(err)
		return
	}

	_, err = w.Write(jsonBytes)
	if err != nil {
		writeJSONError(err)
		return
	}
}

func rawResponse(w http.ResponseWriter, bi interface{}, err error) {
	writeRawError := func(rawErr error) {
		statusCode := he.StatusCodeByError(rawErr)
		w.WriteHeader(statusCode)
		w.Write([]byte(fmt.Sprintf("%s", rawErr)))
		return
	}

	if err != nil {
		writeRawError(err)
		return
	}

	// TODO(sam): this needs to be cleaned up. This expectation in unacceptable.
	b, ok := bi.([]byte)
	if !ok {
		// it is expected that if the handler did not return an object to be
		// converted into JSON, it returned a byte slice.
		writeRawError(errs.New("error casting %+v to bytes", bi))
		return
	}

	_, err = w.Write(b)
	if err != nil {
		writeRawError(err)
		return
	}
}
