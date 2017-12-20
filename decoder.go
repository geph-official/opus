// Copyright © 2015-2017 Go Opus Authors (see AUTHORS file)
//
// License for use of this code is detailed in the LICENSE file

package opus

import (
	"fmt"
	"unsafe"
)

/*
#include <opus/opus.h>
*/
import "C"

var errDecUninitialized = fmt.Errorf("opus decoder uninitialized")

type Decoder struct {
	p *C.struct_OpusDecoder
	// Same purpose as encoder struct
	mem         []byte
	sample_rate int
}

// NewDecoder allocates a new Opus decoder and initializes it with the
// appropriate parameters. All related memory is managed by the Go GC.
func NewDecoder(sample_rate int, channels int) (*Decoder, error) {
	var dec Decoder
	err := dec.Init(sample_rate, channels)
	if err != nil {
		return nil, err
	}
	return &dec, nil
}

func (dec *Decoder) Init(sample_rate int, channels int) error {
	if dec.p != nil {
		return fmt.Errorf("opus decoder already initialized")
	}
	if channels != 1 && channels != 2 {
		return fmt.Errorf("Number of channels must be 1 or 2: %d", channels)
	}
	size := C.opus_decoder_get_size(C.int(channels))
	dec.sample_rate = sample_rate
	dec.mem = make([]byte, size)
	dec.p = (*C.OpusDecoder)(unsafe.Pointer(&dec.mem[0]))
	errno := C.opus_decoder_init(
		dec.p,
		C.opus_int32(sample_rate),
		C.int(channels))
	if errno != 0 {
		return Error(errno)
	}
	return nil
}

// Decode encoded Opus data into the supplied buffer. On success, returns the
// number of samples correctly written to the target buffer.
func (dec *Decoder) Decode(data []byte, pcm []int16, fec bool) (int, error) {
	if dec.p == nil {
		return 0, errDecUninitialized
	}
	if len(pcm) == 0 {
		return 0, fmt.Errorf("opus: target buffer empty")
	}
	var ptr *C.uchar
	if len(data) != 0 {
		ptr = (*C.uchar)(&data[0])
	}
	fc := 0
	if fec {
		fc = 1
	}
	n := int(C.opus_decode(
		dec.p,
		ptr,
		C.opus_int32(len(data)),
		(*C.opus_int16)(&pcm[0]),
		C.int(cap(pcm)),
		C.int(fc)))
	if n < 0 {
		return 0, Error(n)
	}
	return n, nil
}
