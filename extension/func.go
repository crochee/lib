package extension

import (
	"unsafe"

	jsoniter "github.com/json-iterator/go"
)

// Register registers extension to jsoniter.API
func Register(api jsoniter.API) {
	api.RegisterExtension(&u64AsStringCodec{})
}

type funcEncoder struct {
	fun jsoniter.EncoderFunc
}

func (encoder *funcEncoder) Encode(ptr unsafe.Pointer, stream *jsoniter.Stream) {
	encoder.fun(ptr, stream)
}

func (encoder *funcEncoder) IsEmpty(ptr unsafe.Pointer) bool {
	return *((*uint64)(ptr)) == 0
}

type funcDecoder struct {
	fun jsoniter.DecoderFunc
}

func (decoder *funcDecoder) Decode(ptr unsafe.Pointer, iter *jsoniter.Iterator) {
	decoder.fun(ptr, iter)
}
