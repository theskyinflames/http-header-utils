package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/theskyinflames/go-misc/com.theskyinflames.go.misc/lib_gc_event_from_env"
)

const (
	HTTP_HEADER_PROTOCOL_ACCEPT   = "Accept"
	HTTP_HEADER_PROTOCOL_JSON     = "application/json"
	HTTP_HEADER_PROTOCOL_PROTOBUF = "application/x-protobuf"
)

func checkForHttpMethod(method string, r *http.Request, w http.ResponseWriter) error {
	if r.Method != method {
		if msg, err := lib_gc_event_from_env.GetMessageFromEnvTemplate(util.ENV_MESSAGES_PREFIX, "HTTPHEDAERUTILS_001", []string{method, r.Method}); err != nil {
			lib_event_publisher.EventPublisherChannel <- err.Error()
			return err
		} else {
			lib_event_publisher.EventPublisherChannel <- msg
			return errors.New(msg)
		}
	} else {
		return nil
	}
}

func checkForHttpProtocol(requiredProtocol string, r *http.Request) (string, error) {

	protocol := ""
	if protocol = r.Header.Get(util.HTTP_HEADER_PROTOCOL_ACCEPT); protocol == "" {
		param := []string{}
		if msg, err := lib_gc_event_from_env.GetMessageFromEnvTemplate(util.ENV_MESSAGES_PREFIX, "HTTPHEDAERUTILS_002", param); err != nil {
			return "", err
		} else {
			return "", errors.New(msg)
		}
	}

	if strings.Contains(protocol, util.HTTP_HEADER_PROTOCOL_JSON) {
		protocol = util.HTTP_HEADER_PROTOCOL_JSON
	} else if strings.Contains(protocol, util.HTTP_HEADER_PROTOCOL_PROTOBUF) {
		protocol = util.HTTP_HEADER_PROTOCOL_PROTOBUF
	} else {
		param := []string{protocol}
		if msg, err := lib_gc_event_from_env.GetMessageFromEnvTemplate(util.ENV_MESSAGES_PREFIX, "HTTPHEDAERUTILS_003", param); err != nil {
			return "", err
		} else {
			return "", errors.New(msg)
		}
	}

	if protocol != requiredProtocol {
		return "", util.PublishEventMessage("100_009", protocol)
	} else {
		return protocol, nil
	}
}

// The parameter rq_obj is a pointer to the request object whitch the request will be binded to
func getRequestBinded(protocol, rq_obj interface{}, r *http.Request) (interface{}, error) {
	var err error
	if protocol == util.HTTP_HEADER_PROTOCOL_JSON {
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

func returnOK(w http.ResponseWriter, res interface{}, protocol string) error {

	// Serialize the response
	if protocol == util.HTTP_HEADER_PROTOCOL_PROTOBUF {
		if _b, err := proto.Marshal(res.(proto.Message)); err != nil {
			returnError(w, res, err, protocol)
		} else {
			w.Header().Set("Content-type", util.HTTP_HEADER_PROTOCOL_PROTOBUF)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(_b))
		}
	} else if protocol == util.HTTP_HEADER_PROTOCOL_JSON {

		if _b, err := json.Marshal(res); err != nil {
			returnError(w, res, err, protocol)
		} else {
			w.Header().Set("Content-Type", util.HTTP_HEADER_PROTOCOL_JSON)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(_b))
		}
	}

	return nil
}

func returnError(w http.ResponseWriter, res interface{}, _err error, protocol string) error {
	lib_event_publisher.EventPublisherChannel <- _err.Error()

	if protocol == util.HTTP_HEADER_PROTOCOL_PROTOBUF {
		if _b, err := proto.Marshal(res.(proto.Message)); err != nil {
			if msg, err := lib_gc_event_from_env.GetMessageFromEnvTemplate(util.ENV_MESSAGES_PREFIX, "HTTPHEDAERUTILS_004", []string{err.Error()}); err != nil {
				lib_event_publisher.EventPublisherChannel <- err.Error()
			} else {
				lib_event_publisher.EventPublisherChannel <- msg
			}
		} else {
			w.Header().Set("Content-type", util.HTTP_HEADER_PROTOCOL_PROTOBUF)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(_b))
		}
	} else if protocol == util.HTTP_HEADER_PROTOCOL_JSON {

		if _b, err := json.Marshal(res); err != nil {
			if msg, err := lib_gc_event_from_env.GetMessageFromEnvTemplate(util.ENV_MESSAGES_PREFIX, "HTTPHEDAERUTILS_005", []string{err.Error()}); err != nil {
				lib_event_publisher.EventPublisherChannel <- err.Error()
			} else {
				lib_event_publisher.EventPublisherChannel <- msg
			}
		} else {
			w.Header().Set("Content-Type", util.HTTP_HEADER_PROTOCOL_JSON)
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

func unmarshalErrorRs(marshaled string) (*mdhconsolecomparator.ErrorResponse, error) {
	var errorRs mdhconsolecomparator.ErrorResponse
	if err := json.Unmarshal([]byte(marshaled), &errorRs); err != nil {
		return nil, err
	} else {
		return &errorRs, nil
	}
}
