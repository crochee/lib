package lirity

import (
	"io"
	"log"
)

// Close clear ,use defer
func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Println(err)
	}
}
