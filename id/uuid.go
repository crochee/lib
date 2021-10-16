package id

import (
	"github.com/crochee/uid"
	"github.com/satori/go.uuid"
)

func Uuid() string {
	return uuid.NewV4().String()
}

func Uid() string {
	return uid.New().String()
}
