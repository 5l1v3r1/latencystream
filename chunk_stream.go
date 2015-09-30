package latencystream

import "io"

const ReaderMaxChunkSize = 1 << 16

// A ChunkStream is a stream of binary data which arrives in discrete chunks.
// A ChunkStream has no notion of errors; only of completion.
type ChunkStream struct {
	// Chunks is the channel on which chunks are delivered sequentially.
	// It will be closed when the stream is complete.
	Chunks <-chan []byte

	// Close is a stream which, when closed, will signal the stream to stop delivering chunks.
	// It allows the source of the stream to stop blocking while waiting to send output chunks.
	Close chan<- struct{}
}

// NewChunkStreamReader creates a ChunkStream which wraps an io.Reader.
// The size of the chunks will vary depending on the io.Reader.
//
// After calling this, you should not read from the reader manually, even if you close the
// ChunkStream.
func NewChunkStreamReader(r io.Reader) ChunkStream {
	rawInput := make(chan []byte)
	cancel := make(chan struct{})
	go func() {
		defer close(rawInput)
		for {
			buffer := make([]byte, ReaderMaxChunkSize)
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
	return ChunkStream{rawInput, cancel}
}
