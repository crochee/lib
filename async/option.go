package async

import (
	jsoniter "github.com/json-iterator/go"

	"github.com/crochee/lirity/validator"
)

type option struct {
	manager   ManagerCallback
	marshal   MarshalAPI // mq  assemble request or response
	handler   jsoniter.API
	validator validator.Validator
}

type Option func(*option)

func WithManagerCallback(manager ManagerCallback) Option {
	return func(o *option) {
		o.manager = manager
	}
}

func WithMarshalAPI(marshal MarshalAPI) Option {
	return func(o *option) {
		o.marshal = marshal
	}
}

func WithJSON(handler jsoniter.API) Option {
	return func(o *option) {
		o.handler = handler
	}
}

func WithValidator(validator validator.Validator) Option {
	return func(o *option) {
		o.validator = validator
	}
}
