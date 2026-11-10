package channel

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validChannel() Channel {
	return Channel{
		ID:             uuid.New(),
		Number:         1,
		Name:           "El Trece",
		Enabled:        true,
		SegmentSeconds: DefaultSegmentSeconds,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func TestChannel_Validate_OK(t *testing.T) {
	c := validChannel()
	c.Sources[1] = &Source{ID: uuid.New(), Slot: 2, URL: "https://example.com/index.m3u8", Enabled: true}

	require.NoError(t, c.Validate())
}

func TestChannel_Validate_NumberNotPositive(t *testing.T) {
	c := validChannel()
	c.Number = 0

	assertValidationField(t, c.Validate(), "number")
}

func TestChannel_Validate_NameEmpty(t *testing.T) {
	c := validChannel()
	c.Name = ""

	assertValidationField(t, c.Validate(), "name")
}

func TestChannel_Validate_NameTooLong(t *testing.T) {
	c := validChannel()
	c.Name = strings.Repeat("a", 101)

	assertValidationField(t, c.Validate(), "name")
}

func TestChannel_Validate_SegmentSecondsOutOfRange(t *testing.T) {
	for _, seconds := range []int{1, 11} {
		c := validChannel()
		c.SegmentSeconds = seconds

		assertValidationField(t, c.Validate(), "segment_seconds")
	}
}

func TestChannel_Validate_ScheduleURLInvalid(t *testing.T) {
	c := validChannel()
	c.ScheduleURL = "ftp://example.com/schedule"

	assertValidationField(t, c.Validate(), "schedule_url")
}

func TestChannel_Validate_ScheduleURLOptional(t *testing.T) {
	c := validChannel()
	c.ScheduleURL = ""

	require.NoError(t, c.Validate())

	c.ScheduleURL = "https://example.com/schedule"
	require.NoError(t, c.Validate())
}

func TestChannel_Validate_SourceSlotMismatch(t *testing.T) {
	c := validChannel()
	c.Sources[1] = &Source{ID: uuid.New(), Slot: 3, URL: "https://example.com/index.m3u8"}

	assertValidationField(t, c.Validate(), "sources")
}

func TestChannel_Validate_SourceInvalidPropagates(t *testing.T) {
	c := validChannel()
	c.Sources[0] = &Source{ID: uuid.New(), Slot: 1, URL: "not-a-url"}

	assertValidationField(t, c.Validate(), "url")
}

func TestChannel_MarshalJSON_AlwaysFiveSlots(t *testing.T) {
	c := validChannel()
	c.Sources[0] = &Source{ID: uuid.New(), Slot: 1, URL: "https://example.com/index.m3u8", Enabled: true}

	data, err := json.Marshal(c)
	require.NoError(t, err)

	var decoded struct {
		Sources []json.RawMessage `json:"sources"`
	}
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.Len(t, decoded.Sources, SourceSlots)
	assert.JSONEq(t, `null`, string(decoded.Sources[1]))
	assert.JSONEq(t, `null`, string(decoded.Sources[4]))
	assert.NotEqual(t, "null", string(decoded.Sources[0]))
}

func TestNewChannel_Defaults(t *testing.T) {
	c := NewChannel(7, "TV Pública")

	assert.NotEqual(t, uuid.Nil, c.ID)
	assert.Equal(t, 7, c.Number)
	assert.Equal(t, "TV Pública", c.Name)
	assert.True(t, c.Enabled)
	assert.Equal(t, DefaultSegmentSeconds, c.SegmentSeconds)
	assert.False(t, c.CreatedAt.IsZero())
	assert.False(t, c.UpdatedAt.IsZero())
	require.NoError(t, c.Validate())
}

func TestErrValidation_AsAndFields(t *testing.T) {
	err := error(ErrValidation{Field: "name", Msg: "required"})

	var verr ErrValidation
	require.ErrorAs(t, err, &verr)
	assert.Equal(t, "name", verr.Field)
	assert.Equal(t, map[string]string{"name": "required"}, verr.ValidationFields())
	assert.Contains(t, verr.Error(), "name")
}

func TestSentinelErrors_Wrapping(t *testing.T) {
	assert.ErrorIs(t, errors.Join(ErrNotFound), ErrNotFound)
	assert.ErrorIs(t, errors.Join(ErrConflict), ErrConflict)
	assert.NotErrorIs(t, ErrNotFound, ErrConflict)
}

func assertValidationField(t *testing.T, err error, field string) {
	t.Helper()

	require.Error(t, err)
	var verr ErrValidation
	require.ErrorAs(t, err, &verr)
	assert.Equal(t, field, verr.Field)
}
