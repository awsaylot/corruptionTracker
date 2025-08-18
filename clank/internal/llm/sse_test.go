package llm

import (
	"bufio"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessSSEStream(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedChunks []string
		expectError    bool
	}{
		{
			name: "successful stream processing",
			input: `data: First chunk
event: message

data: Second chunk
event: message

data: Final chunk
event: message

`,
			expectedChunks: []string{
				"First chunk",
				"Second chunk",
				"Final chunk",
			},
			expectError: false,
		},
		{
			name: "stream with done event",
			input: `data: First chunk
event: message

data: Second chunk
event: message

event: done

`,
			expectedChunks: []string{
				"First chunk",
				"Second chunk",
			},
			expectError: false,
		},
		{
			name: "stream with error event",
			input: `data: First chunk
event: message

event: error
data: Something went wrong

`,
			expectedChunks: []string{
				"First chunk",
			},
			expectError: true,
		},
		{
			name: "malformed stream",
			input: `invalid
format
data`,
			expectError: true,
		},
		{
			name:        "empty stream",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			chunks := make(chan string, len(tt.expectedChunks))
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			err := ProcessSSEStream(ctx, reader, chunks)
			close(chunks)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			var receivedChunks []string
			for chunk := range chunks {
				receivedChunks = append(receivedChunks, chunk)
			}

			assert.Equal(t, tt.expectedChunks, receivedChunks)
		})
	}
}

func TestParseSSEEvent(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedEvent *SSEEvent
		expectError   bool
	}{
		{
			name: "valid message event",
			input: `event: message
data: {"text": "Hello"}

`,
			expectedEvent: &SSEEvent{
				Event: "message",
				Data:  `{"text": "Hello"}`,
			},
			expectError: false,
		},
		{
			name: "valid done event",
			input: `event: done
data: Stream completed

`,
			expectedEvent: &SSEEvent{
				Event: "done",
				Data:  "Stream completed",
			},
			expectError: false,
		},
		{
			name: "valid error event",
			input: `event: error
data: Something went wrong

`,
			expectedEvent: &SSEEvent{
				Event: "error",
				Data:  "Something went wrong",
			},
			expectError: false,
		},
		{
			name:        "invalid format",
			input:       "invalid\nformat\n\n",
			expectError: true,
		},
		{
			name:        "empty input",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			event, err := ParseSSEEvent(reader)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expectedEvent.Event, event.Event)
			assert.Equal(t, tt.expectedEvent.Data, event.Data)
		})
	}
}

func TestReadSSELine(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "single line",
			input:       "test line\n",
			expected:    "test line",
			expectError: false,
		},
		{
			name:        "line with carriage return",
			input:       "test line\r\n",
			expected:    "test line",
			expectError: false,
		},
		{
			name:        "empty line",
			input:       "\n",
			expected:    "",
			expectError: false,
		},
		{
			name:        "no newline",
			input:       "test",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(strings.NewReader(tt.input))
			line, err := ReadSSELine(reader)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, line)
		})
	}
}

func TestHandleSSETimeout(t *testing.T) {
	tests := []struct {
		name        string
		timeout     time.Duration
		input       func() io.Reader
		expectError bool
	}{
		{
			name:    "completes before timeout",
			timeout: time.Second,
			input: func() io.Reader {
				return strings.NewReader("data: test\n\n")
			},
			expectError: false,
		},
		{
			name:    "times out",
			timeout: time.Millisecond * 50,
			input: func() io.Reader {
				// Create a slow reader that will trigger timeout
				return &SlowReader{
					data:  []byte("data: test\n\n"),
					delay: time.Millisecond * 100,
				}
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := bufio.NewReader(tt.input())
			chunks := make(chan string, 1)

			ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
			defer cancel()

			err := ProcessSSEStream(ctx, reader, chunks)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "context deadline exceeded")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// SlowReader implements io.Reader for testing timeouts
type SlowReader struct {
	data  []byte
	pos   int
	delay time.Duration
}

func (r *SlowReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	time.Sleep(r.delay)

	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
