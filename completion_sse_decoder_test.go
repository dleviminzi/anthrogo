package anthrogo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		wantEv  *CompletionEvent
	}{
		{
			name:    "has : prefix",
			input:   ":\n",
			wantErr: false,
			wantEv:  nil,
		},
		{
			name:    "carriage w/ no event",
			input:   "\r\n",
			wantErr: false,
			wantEv:  nil,
		},
		{
			name:    "empty line",
			input:   "\n",
			wantErr: false,
			wantEv:  nil,
		},
		{
			name:    "invalid field",
			input:   "some junk data\n",
			wantErr: false,
			wantEv:  nil,
		},
		{
			name:    "id field",
			input:   "id: testID\n\r",
			wantErr: false,
			wantEv:  &CompletionEvent{ID: "testID"},
		},
		{
			name:    "event field",
			input:   "event: testEvent\n\r",
			wantErr: false,
			wantEv:  &CompletionEvent{Event: "testEvent"},
		},
		{
			name:    "retry field",
			input:   "retry: 5\n\r",
			wantErr: false,
			wantEv:  &CompletionEvent{Retry: 5},
		},
		{
			name:    "data field",
			input:   "data: {\"completion\":\"testCompletion\",\"stop_reason\":\"testReason\",\"model\":\"testModel\",\"stop\":\"testStop\",\"log_id\":\"testLogId\"}\n\r",
			wantErr: false,
			wantEv: &CompletionEvent{
				Data: &CompletionEventData{
					Completion: "testCompletion",
					StopReason: "testReason",
					Model:      "testModel",
					Stop:       "testStop",
					LogID:      "testLogId",
				},
			},
		},
		{
			name:    "invalid json in data field",
			input:   "data: {\"completion\":\"testCompletion\",}",
			wantErr: true,
		},
		{
			name:    "invalid integer in retry field",
			input:   "retry: invalid\n",
			wantErr: false,
			wantEv:  nil,
		},
		{
			name:    "id field with null byte",
			input:   "id: test\000ID\n",
			wantErr: false,
			wantEv:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input + "\n")
			dec := NewCompletionSSEDecoder(r)

			ev, err := dec.Decode()
			if !tt.wantErr {
				ev, err = dec.Decode()
			}

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.wantEv, ev, "they should be equal")
		})
	}
}
