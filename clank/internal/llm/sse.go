package llm

import (
	"bufio"
	"bytes"
	"io"
	"strings"
)

// SSEEvent represents a Server-Sent Event
type SSEEvent struct {
	Event string
	Data  string
}

// SSEReader reads Server-Sent Events from a stream
type SSEReader struct {
	reader *bufio.Reader
}

// NewSSEReader creates a new SSE reader from an io.Reader
func NewSSEReader(r io.Reader) *SSEReader {
	return &SSEReader{
		reader: bufio.NewReader(r),
	}
}

// Read reads a single SSE event
func (r *SSEReader) Read() (*SSEEvent, error) {
	event := &SSEEvent{}
	buffer := bytes.Buffer{}

	for {
		line, err := r.reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				// If we have data in the buffer, process it before returning EOF
				if buffer.Len() > 0 {
					event.Data = strings.TrimSpace(buffer.String())
					return event, nil
				}
			}
			return nil, err
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			// Empty line indicates the end of an event
			if buffer.Len() > 0 {
				event.Data = strings.TrimSpace(buffer.String())
				return event, nil
			}
			continue
		}

		switch {
		case bytes.HasPrefix(line, []byte("event:")):
			event.Event = string(bytes.TrimSpace(bytes.TrimPrefix(line, []byte("event:"))))
		case bytes.HasPrefix(line, []byte("data:")):
			data := bytes.TrimSpace(bytes.TrimPrefix(line, []byte("data:")))
			if len(data) > 0 {
				buffer.Write(data)
				buffer.WriteByte('\n')
			}
		case bytes.Equal(line, []byte("[DONE]")):
			return nil, io.EOF
		}
	}
}
