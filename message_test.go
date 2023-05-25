package connqc_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncoder_Encode(t *testing.T) {
	tests := []struct {
		name      string
		msg       signal.Message
		wantBytes []byte
		wantErr   require.ErrorAssertionFunc
	}{
		{
			name:      "handles encoding probe",
			msg:       signal.Probe{ID: 2, Data: "Hello 2"},
			wantBytes: []byte{'P', 'R', 'B', 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x7, 'H', 'e', 'l', 'l', 'o', ' ', '2'},
			wantErr:   require.NoError,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			enc := signal.NewEncoder(&buf)

			err := enc.Encode(test.msg)

			test.wantErr(t, err)
			assert.Equal(t, test.wantBytes, buf.Bytes())
		})
	}
}

func TestDecoder_Decode(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantMsg signal.Message
		wantErr require.ErrorAssertionFunc
	}{
		{
			name:    "handles encoding probe",
			data:    []byte{'P', 'R', 'B', 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x2, 0x0, 0x7, 'H', 'e', 'l', 'l', 'o', ' ', '2'},
			wantMsg: signal.Probe{ID: 2, Data: "Hello 2"},
			wantErr: require.NoError,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			dec := signal.NewDecoder(bytes.NewReader(test.data))

			got, err := dec.Decode()

			test.wantErr(t, err)
			assert.Equal(t, test.wantMsg, got)
		})
	}
}
