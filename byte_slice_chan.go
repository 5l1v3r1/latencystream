package latencyreader

import "io"

const ReadBufferSize = 1 << 16

// A ByteSliceChan represents a stream of binary data in the form of a channel.
// The stream is "done" when the channel is closed.
// A BufferChan has no notion of errors; only of "done-ness".
type ByteSliceChan <-chan []byte

// NewBufferChanReader creates a BufferChan which wraps an io.Reader.
// The size of the output byte slices will be determined in part by the specific io.Reader.
func NewByteSliceChanReader(r io.Reader) ByteSliceChan {
	rawInput := make(chan []byte)
	go func() {
		for {
			buffer := make([]byte, ReadBufferSize)
			count, err := r.Read(buffer)
			if count > 0 {
				rawInput <- buffer[:count]
			} else if err != nil {
				break
			}
		}
		close(rawInput)
	}()
	return rawInput
}
