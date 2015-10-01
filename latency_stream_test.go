package latencystream

import (
	"bytes"
	"testing"
	"time"
)

func TestNewLatencyStreamInfLatency(t *testing.T) {
	unbuffered := make(chan []byte)
	closeChan := make(chan struct{})
	chunkStream := ChunkStream{unbuffered, closeChan}
	latencyStream := NewLatencyStream(chunkStream, 10, time.Hour)

	unbuffered <- []byte("hey")
	unbuffered <- []byte("testing")
	valueOut := <-latencyStream.Chunks
	if !bytes.Equal(valueOut, []byte("heytesting")) {
		t.Error("bad chunk")
	}

	unbuffered <- []byte("hey")
	unbuffered <- []byte("testing123")
	valueOut = <-latencyStream.Chunks
	if !bytes.Equal(valueOut, []byte("heytesting")) {
		t.Error("bad chunk")
	}

	unbuffered <- []byte("testing")
	valueOut = <-latencyStream.Chunks
	if !bytes.Equal(valueOut, []byte("123testing")) {
		t.Error("bad chunk")
	}

	close(latencyStream.Close)
	time.Sleep(time.Millisecond * 100)
	select {
	case <-closeChan:
	default:
		t.Error("closing did not result in more closes")
	}
}

func TestNewLatencyStream(t *testing.T) {
	// TODO: this.
	t.Error("not yet implemented")
}

