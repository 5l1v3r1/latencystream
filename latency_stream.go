package latencystream

import "time"

// NewLatencyStream creates a ChunkStream which buffers another ChunkStream.
//
// The returned stream will send fixed-sized chunks unless a maximum input latency is exceeded, in
// which case smaller chunks may be sent.
// In this case, latency is defined as the amount of time that the oldest data in the buffer has
// been waiting.
func NewLatencyStream(source ChunkStream, chunkSize int, maxLatency time.Duration) ChunkStream {
	buf := newDataBuffer(source, chunkSize, maxLatency)
	go func() {
		buf.runLoop()
	}()
	return buf.outputChunkStream()
}

type dataBuffer struct {
	buffer     []byte
	maxChunk   int
	maxLatency time.Duration

	// timeout gets a message when the oldest data in the buffer has been there for longer than
	// MaxLatency amount of time.
	timeout <-chan time.Time

	unlimitedTimeout <-chan time.Time

	input ChunkStream

	output chan []byte
	closer chan struct{}
}

func newDataBuffer(source ChunkStream, maxChunk int, maxLatency time.Duration) *dataBuffer {
	outputChan := make(chan []byte)
	closerChan := make(chan struct{})
	return &dataBuffer{maxChunk: maxChunk, maxLatency: maxLatency,
		unlimitedTimeout: make(chan time.Time), input: source, output: outputChan,
		closer: closerChan}
}

func (b *dataBuffer) outputChunkStream() ChunkStream {
	return ChunkStream{b.output, b.closer}
}

func (b *dataBuffer) runLoop() {
	defer close(b.output)
	b.resetTimeout()
	for {
		select {
		case <-b.closer:
			close(b.input.Close)
			return
		default:
		}

		select {
		case inputData := <-b.input.Chunks:
			if inputData == nil {
				b.finish()
				return
			}

			if len(b.buffer) == 0 {
				b.resetTimeout()
			}

			b.buffer = append(b.buffer, inputData...)
			if b.hasFullChunk() {
				b.resetTimeout()
				b.sendFullChunks()
			}
		case <-b.timeout:
			b.handleTimeout()
		case <-b.closer:
			return
		}
	}
}

func (b *dataBuffer) finish() {
	if b.hasFullChunk() {
		b.sendFullChunks()
	}
	b.output <- b.buffer
}

func (b *dataBuffer) handleTimeout() {
	if len(b.buffer) == 0 {
		b.setUnlimitedTimeout()
		return
	}

	for {
		select {
		case b.output <- b.buffer:
			b.buffer = make([]byte, 0)
			b.setUnlimitedTimeout()
			return
		case moreData := <-b.input.Chunks:
			b.buffer = append(b.buffer, moreData...)
			if b.hasFullChunk() {
				b.resetTimeout()
				b.sendFullChunks()
				return
			}
		case <-b.closer:
			return
		}
	}
}

func (b *dataBuffer) sendFullChunks() {
	for b.hasFullChunk() {
		outBuf := make([]byte, b.maxChunk)
		copy(outBuf, b.buffer)

		select {
		case b.output <- outBuf:
		case <-b.closer:
			return
		}

		copy(b.buffer, b.buffer[b.maxChunk:])
		b.buffer = b.buffer[:len(b.buffer)-b.maxChunk]
	}
}

func (b *dataBuffer) hasFullChunk() bool {
	return len(b.buffer) >= b.maxChunk
}

func (b *dataBuffer) resetTimeout() {
	b.timeout = time.After(b.maxLatency)
}

func (b *dataBuffer) setUnlimitedTimeout() {
	b.timeout = b.unlimitedTimeout
}
