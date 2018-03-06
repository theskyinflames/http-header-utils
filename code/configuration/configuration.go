package configuration

import (
	"errors"

	"github.com/theskyinflames/go-misc/com.theskyinflames.go.misc/lib_gc_event_from_env"
	"github.com/theskyinflames/go-misc/com.theskyinflames.go.misc/lib_gc_event_publisher"
	UTIL "github.com/theskyinflames/go-misc/com.theskyinflames.go.misc/lib_gc_util"
)

func init() {
	_ENV_MESSAGES_PREFIX := &UTIL.EnvironmentVariable{Var_name: "GO_HTTPHEADERUTILS_PREFIX", Var_type: UTIL.ENVIRONMENT_VARIABLE_TYPE_STRING}
	evs := []*UTIL.EnvironmentVariable{_ENV_MESSAGES_PREFIX}
	if _, err := UTIL.GetEnvironmentVariables(evs); err != nil {
		panic(err)
	} else {
		ENV_MESSAGES_PREFIX = _ENV_MESSAGES_PREFIX.Var_value.(string)
	}
}

// This variable is used as prefix for the environment variables
// wich are used to store app messages. For example, it may be
// to have an environment variable like that:
//     GO_APP_MSG01="This is he message 01"
// If the prefix is defined as "GO_APP", this message can be
// retrieved by its code MSG01.
// See github.com/theskyinflames/go-misc/com.theskyinflames.go.misc/lib_gc_event_from_env
// For more detail
var ENV_MESSAGES_PREFIX string

func PublishEventMessage(code string, params ...string) error {
	if msg, err := lib_gc_event_from_env.GetMessageFromEnvTemplate(ENV_MESSAGES_PREFIX, code, params); err != nil {
		lib_gc_event_publisher.EventPublisherChannel <- err.Error()
		return err
	} else {
		lib_gc_event_publisher.EventPublisherChannel <- msg
		return errors.New(msg)
	}
}
