package id

import (
	"github.com/crochee/uid"
	"github.com/satori/go.uuid"
)

func UUID() string {
	return uuid.NewV4().String()
}

func UID() string {
	return uid.New().String()
}
