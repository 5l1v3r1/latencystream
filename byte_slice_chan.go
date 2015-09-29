package latencyreader

import "io"

const ReadBufferSize = 1 << 16

// A ByteSliceChan represents a stream of binary data in the form of a channel.
// The stream is "done" when the channel is closed.
// A BufferChan has no notion of errors; only of "done-ness".
type ByteSliceChan <-chan []byte

// NewByteSliceChanReader creates a ByteSliceChan which wraps an io.Reader.
// The size of the output byte slices will be determined in part by the specific io.Reader.
//
// You may close the cancelChan to finish the ByteSliceChan early.
// Closing the cancelChan does not guarantee that no more reads will be performed.
// After creating the ByteSliceChan, you should never read from the reader manually again.
func NewByteSliceChanReader(r io.Reader) (c ByteSliceChan, cancelChan chan<- struct{}) {
	rawInput := make(chan []byte)
	cancel := make(chan struct{})
	go func() {
		defer close(rawInput)
		for {
			buffer := make([]byte, ReadBufferSize)
			count, err := r.Read(buffer)
			if count > 0 {
				// NOTE: if someone is reading our output channel but we are cancelled, there is
				// probabliity 1/2^n that outputs will be sent rather than reading the cancelChan.
				// To address this, we first check for cancelChan before doing the second select{}.
				select {
				case <-cancel:
					return
				default:
				}

				select {
				case rawInput <- buffer[:count]:
				case <-cancel:
					return
				}
			}
			if err != nil {
				return
			}
		}
		close(rawInput)
	}()
	return rawInput, cancel
}
