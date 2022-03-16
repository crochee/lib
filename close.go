package lirity

import (
	"io"
	"log"
)

// Close clear ,use defer
func Close(c io.Closer) {
	if c == nil {
		return
	}
	if err := c.Close(); err != nil {
		log.Println(err)
	}
}
