package id

import (
	"github.com/satori/go.uuid"
)

func UV4() string {
	return uuid.NewV4().String()
}
