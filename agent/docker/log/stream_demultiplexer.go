package log

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/text/transform"
)

// streamDemultiplexer removes header from multiplexed stdio stream returned from Docker Engine API.
// It implements transform.Transformer.
type streamDemultiplexer struct {
	remainingPayload int
}

var _ transform.Transformer = (*streamDemultiplexer)(nil)

func newStreamDemultiplexer() *streamDemultiplexer {
	return &streamDemultiplexer{}
}

// Reset resets the state and allows a streamDemultiplexer to be reused.
// It implements transform.Transformer.Reset.
func (s *streamDemultiplexer) Reset() {
	s.remainingPayload = 0
}

// Transform copies payload from src to dst.
// It implements transform.Transformer.Transform.
func (s *streamDemultiplexer) Transform(dst, src []byte, atEOF bool) (nDst int, nSrc int, err error) {
	var totalWrittenBytes, totalReadBytes int

	for totalReadBytes < len(src) {
		writtenBytes, readBytes, err := s.partialTransform(dst[totalWrittenBytes:], src[totalReadBytes:], atEOF)

		totalWrittenBytes += writtenBytes
		totalReadBytes += readBytes

		if err != nil {
			return totalWrittenBytes, totalReadBytes, err
		}
	}

	return totalWrittenBytes, totalReadBytes, nil
}

func (s *streamDemultiplexer) partialTransform(dst, src []byte, atEOF bool) (advanceDst int, advanceSrc int, err error) {
	if atEOF && len(src) == 0 {
		return 0, 0, nil
	}

	if s.remainingPayload == 0 {
		headerSize, payloadSize, err := parseHeader(src)
		if err != nil {
			return 0, 0, err
		}

		s.remainingPayload = payloadSize
		return 0, headerSize, nil
	}

	// Read payload
	bytesToConsume := s.remainingPayload
	if len(src) < bytesToConsume {
		bytesToConsume = len(src)
	}

	consumedBytes := copy(dst, src[:bytesToConsume])
	s.remainingPayload -= consumedBytes

	if consumedBytes < bytesToConsume {
		return consumedBytes, consumedBytes, transform.ErrShortDst
	} else {
		return consumedBytes, consumedBytes, nil
	}
}

func parseHeader(src []byte) (headerSize int, payloadSize int, err error) {
	if len(src) == 0 {
		// Request more data to parse header.
		return 0, 0, transform.ErrShortSrc
	}

	if src[0] > 2 {
		// Stream type must be 0 (stdin), 1 (stdout), or 2 (stderr).
		return 0, 0, fmt.Errorf("unknown stream type: %d", src[0])
	}

	if len(src) < 8 {
		// Request more data to parse header.
		return 0, 0, transform.ErrShortSrc
	}

	return 8, int(binary.BigEndian.Uint32(src[4:8])), nil
}
