package latencyreader

import (
	"io"
	"time"
)

const MaxLatency = time.Second / 2
const MaxBuffer = 1 << 20

// BufferReader turns an io.Reader into a channel of buffered data chunks.
func BufferReader(r io.Reader) <-chan []byte {
	output := make(chan []byte)
	go func() {
		defer close(output)
		chunks := NewChunkStreamReader(r)
		rawInput := chunks.Chunks
		buffer := &DataBufferer{MaxBuffer: MaxBuffer, MaxLatency: MaxLatency, Input: rawInput,
			Output: output}
		buffer.Run()
	}()
	return output
}

type DataBufferer struct {
	Buffer     []byte
	MaxBuffer  int
	MaxLatency time.Duration

	// Timeout gets a message when the oldest data in the buffer has been there for longer than
	// MaxLatency.
	Timeout <-chan time.Time

	Input  <-chan []byte
	Output chan<- []byte
}

// Run buffers data from the input and sends it to the output.
// It returns once the input has been closed and all data is finished being written to the output.
func (b *DataBufferer) Run() {
	b.resetTimeout()
	for {
		select {
		case inputData := <-b.Input:
			if inputData == nil {
				b.finish()
				return
			}
			b.append(inputData)
			if b.isFull() {
				b.resetTimeout()
				b.sendFullPackets()
			}
		case <-b.Timeout:
			b.sendLatentData()
		}
	}
}

// sendFullPackets sends one or more packets which are each b.MaxBuffer bytes long.
// It will not update the channel timeout.
func (b *DataBufferer) sendFullPackets() {
	if !b.isFull() {
		panic("not enough data to send full packet")
	}

	for b.isFull() {
		outBuf := make([]byte, b.MaxBuffer)
		copy(outBuf, b.Buffer)
		b.Output <- outBuf

		copy(b.Buffer, b.Buffer[b.MaxBuffer:])
		b.Buffer = b.Buffer[:len(b.Buffer)-b.MaxBuffer]
	}
}

// sendLatentData sends a packet or set of packets at the first possible time.
// It accumulates incoming data as it waits for the output channel to unblock.
// It resets b.Timeout appropriately.
func (b *DataBufferer) sendLatentData() {
	if len(b.Buffer) == 0 {
		b.resetTimeout()
		return
	}

	for {
		select {
		case b.Output <- b.Buffer:
			b.Buffer = make([]byte, 0)
			b.resetTimeout()
			return
		case moreData := <-b.Input:
			b.append(moreData)
			if b.isFull() {
				b.resetTimeout()
				b.sendFullPackets()
				return
			}
		}
	}
}

// finish sends all of the remaining data from the buffer to the output.
func (b *DataBufferer) finish() {
	if b.isFull() {
		b.sendFullPackets()
	}
	b.Output <- b.Buffer
}

func (b *DataBufferer) append(d []byte) {
	b.Buffer = append(b.Buffer, d...)
}

func (b *DataBufferer) isFull() bool {
	return len(b.Buffer) > b.MaxBuffer
}

func (b *DataBufferer) resetTimeout() {
	b.Timeout = time.After(b.MaxLatency)
}
