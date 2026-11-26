package tvchannels_test

import (
	"testing"
	"time"

	"github.com/rodriez/RodriTV/internal/tvchannels"
	"github.com/stretchr/testify/suite"
)

type TvChannelTestSuite struct {
	suite.Suite
}

// Init the suite
func (s *TvChannelTestSuite) SetupTest() {}

func (s *TvChannelTestSuite) completeTvChannel() tvchannels.TvChannel {
	return tvchannels.TvChannel{
		ID:           "22fa6d53-6e88-42ae-8721-64c81f6a2632",
		Name:         "Telefe",
		ListPosition: 1,
		Status:       tvchannels.TVChannelStatusActive,
		Signals: tvchannels.TvSignalSet{
			{
				URL:         "https://www.mitelefe.com/telefe-en-vivo",
				Priority:    1,
				NeedBrowser: false,
			},
			{
				URL:         "http://45.134.141.161:2200/ARG/TELEFE_HD/index.m3u8",
				Priority:    2,
				NeedBrowser: false,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *TvChannelTestSuite) TestInvalidName() {
	tvchannel := s.completeTvChannel()
	tvchannel.Name = ""

	err := tvchannel.Validate()

	s.EqualError(err, tvchannels.ErrInvalidName.Error())
}

func (s *TvChannelTestSuite) TestInvalidListPosition() {
	tvchannel := s.completeTvChannel()
	tvchannel.ListPosition = -1

	err := tvchannel.Validate()

	s.EqualError(err, tvchannels.ErrInvalidListPosition.Error())
}

func (s *TvChannelTestSuite) TestInvalidStatus() {
	tvchannel := s.completeTvChannel()
	tvchannel.Status = tvchannels.TVChannelStatus("FakeStatus")

	err := tvchannel.Validate()

	s.EqualError(err, tvchannels.ErrInvalidStatus.Error())
}

func (s *TvChannelTestSuite) TestEmptySignals() {
	tvchannel := s.completeTvChannel()
	tvchannel.Signals = tvchannels.TvSignalSet{}

	err := tvchannel.Validate()

	s.EqualError(err, tvchannels.ErrEmptySignals.Error())
}

func (s *TvChannelTestSuite) TestInvalidSignalURL() {
	tvchannel := s.completeTvChannel()
	tvchannel.Signals[1].URL = ""

	err := tvchannel.Validate()

	s.EqualError(err, tvchannels.ErrInvalidSignalURL.Error())
}

func (s *TvChannelTestSuite) TestInvalidSignalPriority() {
	tvchannel := s.completeTvChannel()
	tvchannel.Signals[1].Priority = -1

	err := tvchannel.Validate()

	s.EqualError(err, tvchannels.ErrInvalidSignalPriority.Error())
}

func TestTvChannel(t *testing.T) {
	suite.Run(t, new(TvChannelTestSuite))
}
