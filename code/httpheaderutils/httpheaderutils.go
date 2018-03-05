package httpheaderutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/theskyinflames/http-header-utils/code/configuration"
)

// For compilation purposes
var Dummy struct{}

const (
	HTTP_HEADER_PROTOCOL_ACCEPT   = "Accept"
	HTTP_HEADER_PROTOCOL_JSON     = "application/json"
	HTTP_HEADER_PROTOCOL_PROTOBUF = "application/x-protobuf"
)

// It validates that the incoming method is the expected one
func CheckForHttpMethod(method string, r *http.Request) error {
	if r.Method != method {
		return configuration.PublishEventMessage("HTTPHEDAERUTILS_001", method)
	} else {
		return nil
	}
}

// It checks for an incomming header "Accept" filled whith the required
// protocol: "application/json" or "application/x-protobuf". If the header
// does not exists, it will result an error. If exist, que rq protocol is
// set from its value.
func CheckForHttpProtocol(requiredProtocol string, r *http.Request) (string, error) {

	protocol := ""
	if protocol = r.Header.Get(HTTP_HEADER_PROTOCOL_ACCEPT); protocol == "" {
		return "", configuration.PublishEventMessage("HTTPHEDAERUTILS_002")
	}

	if strings.Contains(protocol, HTTP_HEADER_PROTOCOL_JSON) {
		protocol = HTTP_HEADER_PROTOCOL_JSON
	} else if strings.Contains(protocol, HTTP_HEADER_PROTOCOL_PROTOBUF) {
		protocol = HTTP_HEADER_PROTOCOL_PROTOBUF
	} else {
		return "", configuration.PublishEventMessage("HTTPHEDAERUTILS_003", protocol)
	}

	if protocol != requiredProtocol {
		return "", configuration.PublishEventMessage("HTTPHEDAERUTILS_004", protocol)
	} else {
		return protocol, nil
	}
}

// The parameter rq_obj is a pointer to the request object whitch the request will be binded to
func GetRequestBinded(protocol, rq_obj interface{}, r *http.Request) (interface{}, error) {
	var err error
	if protocol == HTTP_HEADER_PROTOCOL_JSON {
		if err = json.NewDecoder(r.Body).Decode(rq_obj); err != nil {
			return nil, err
		}
	} else {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		if err = proto.Unmarshal(buf.Bytes(), rq_obj.(proto.Message)); err != nil {
			return nil, err
		}
	}
	return rq_obj, nil
}

// It returns an OK response
func ReturnOK(w http.ResponseWriter, res interface{}, protocol string) error {

	// Serialize the response
	if protocol == HTTP_HEADER_PROTOCOL_PROTOBUF {
		if _b, err := proto.Marshal(res.(proto.Message)); err != nil {
			ReturnError(w, res, err, protocol)
		} else {
			w.Header().Set("Content-type", HTTP_HEADER_PROTOCOL_PROTOBUF)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(_b))
		}
	} else if protocol == HTTP_HEADER_PROTOCOL_JSON {

		if _b, err := json.Marshal(res); err != nil {
			ReturnError(w, res, err, protocol)
		} else {
			w.Header().Set("Content-Type", HTTP_HEADER_PROTOCOL_JSON)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(_b))
		}
	}

	return nil
}

// It returns an ERROR response
func ReturnError(w http.ResponseWriter, res interface{}, _err error, protocol string) error {

	if protocol == HTTP_HEADER_PROTOCOL_PROTOBUF {
		if _b, err := proto.Marshal(res.(proto.Message)); err != nil {
			return configuration.PublishEventMessage("HTTPHEDAERUTILS_005", err.Error())
		} else {
			w.Header().Set("Content-type", HTTP_HEADER_PROTOCOL_PROTOBUF)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(_b))
		}
	} else if protocol == HTTP_HEADER_PROTOCOL_JSON {

		if _b, err := json.Marshal(res); err != nil {
			return configuration.PublishEventMessage("HTTPHEDAERUTILS_006", err.Error())
		} else {
			w.Header().Set("Content-Type", HTTP_HEADER_PROTOCOL_JSON)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(_b))
		}
	} else {

		// HTTP protocol not supported
		w.WriteHeader(http.StatusInternalServerError)
		http.Error(w, fmt.Sprintf("HTTP protocol [%s] not supported ", protocol), http.StatusInternalServerError)
	}

	return nil
}
