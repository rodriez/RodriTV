package channel

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSource_Validate_OK(t *testing.T) {
	s := Source{ID: uuid.New(), Slot: 1, URL: "http://example.com/stream.m3u8"}

	require.NoError(t, s.Validate())

	s = Source{ID: uuid.New(), Slot: 5, URL: "https://example.com/stream.m3u8"}
	require.NoError(t, s.Validate())
}

func TestSource_Validate_SlotOutOfRange(t *testing.T) {
	for _, slot := range []int{0, 6} {
		s := Source{ID: uuid.New(), Slot: slot, URL: "https://example.com/stream.m3u8"}

		assertValidationField(t, s.Validate(), "slot")
	}
}

func TestSource_Validate_URLNotHTTP(t *testing.T) {
	for _, rawURL := range []string{"", "not-a-url", "ftp://example.com/x", "https:///missing-host"} {
		s := Source{ID: uuid.New(), Slot: 1, URL: rawURL}

		assertValidationField(t, s.Validate(), "url")
	}
}
