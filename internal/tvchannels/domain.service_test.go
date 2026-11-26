package tvchannels_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rodriez/RodriTV/internal/tvchannels"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type ServiceTestSuite struct {
	suite.Suite

	ctrl  *gomock.Controller
	store *MockTvChannelStoreGateway
}

func (s *ServiceTestSuite) SetupTest() {
	s.ctrl = gomock.NewController(s.T())
	s.store = NewMockTvChannelStoreGateway(s.ctrl)
}

func (s *ServiceTestSuite) validAddChannelRequest() tvchannels.TvChannel {
	return tvchannels.TvChannel{
		Name:         "Telefe",
		ListPosition: 1,
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
	}
}

func (s *ServiceTestSuite) existingTvChannel() tvchannels.TvChannel {
	tvchannel := s.validAddChannelRequest()
	tvchannel.ID = "3b871028-be80-4257-8759-53d19f7fbd92"
	tvchannel.Status = tvchannels.TVChannelStatusActive
	tvchannel.CreatedAt = time.Now()
	tvchannel.UpdatedAt = time.Now()

	return tvchannel
}

func (s *ServiceTestSuite) TestAddChannelWithInvalidChannel() {
	candidate := s.validAddChannelRequest()
	candidate.Name = ""

	ctx := context.TODO()
	srv := tvchannels.NewService(s.store)
	_, err := srv.AddChannel(ctx, candidate)

	s.EqualError(err, tvchannels.ErrInvalidName.Error())
}

func (s *ServiceTestSuite) TestAddChannelWithStoreGatewayError() {
	candidate := s.validAddChannelRequest()
	ctx := context.TODO()

	s.store.EXPECT().Create(ctx, gomock.Any()).Return(errors.New("Oops something went wrong!!!"))
	srv := tvchannels.NewService(s.store)
	_, err := srv.AddChannel(ctx, candidate)

	s.EqualError(err, tvchannels.ErrInternal.Error())
}

func (s *ServiceTestSuite) TestAddChannelSuccessfully() {
	candidate := s.validAddChannelRequest()
	ctx := context.TODO()

	s.store.EXPECT().Create(ctx, gomock.Any()).Return(nil)
	srv := tvchannels.NewService(s.store)
	createdChannel, err := srv.AddChannel(ctx, candidate)

	s.Empty(err)
	s.Equal(createdChannel.Name, candidate.Name)
	s.Equal(createdChannel.ListPosition, candidate.ListPosition)
	s.Equal(createdChannel.Signals, candidate.Signals)
	s.Equal(createdChannel.Status, tvchannels.TVChannelStatusPaused)
}

func (s *ServiceTestSuite) TestGetChannelWithEmptyId() {
	ctx := context.TODO()
	srv := tvchannels.NewService(s.store)
	_, err := srv.GetChannel(ctx, "")

	s.EqualError(err, tvchannels.ErrInvalidID.Error())
}

func (s *ServiceTestSuite) TestGetChannelWithStoreGatewayError() {
	ctx := context.TODO()

	tvchannelID := "3b871028-be80-4257-8759-53d19f7fbd92"

	s.store.EXPECT().GetByID(ctx, tvchannelID).Return(tvchannels.TvChannel{}, errors.New("Oops something went wrong!!!"))
	srv := tvchannels.NewService(s.store)
	_, err := srv.GetChannel(ctx, tvchannelID)

	s.EqualError(err, tvchannels.ErrInternal.Error())
}

func (s *ServiceTestSuite) TestGetChannelWithServiceEmptyResult() {
	ctx := context.TODO()

	tvchannelID := "3b871028-be80-4257-8759-53d19f7fbd92"

	s.store.EXPECT().GetByID(ctx, tvchannelID).Return(tvchannels.TvChannel{}, nil)
	srv := tvchannels.NewService(s.store)
	_, err := srv.GetChannel(ctx, tvchannelID)

	s.EqualError(err, tvchannels.ErrInvalidID.Error())
}

func (s *ServiceTestSuite) TestGetChannelSuccessfully() {
	existing := s.existingTvChannel()
	ctx := context.TODO()

	s.store.EXPECT().GetByID(ctx, existing.ID).Return(existing, nil)
	srv := tvchannels.NewService(s.store)
	result, err := srv.GetChannel(ctx, existing.ID)

	s.Empty(err)
	s.Equal(result, existing)
}

func (s *ServiceTestSuite) TestReplaceChannelWithEmptyId() {
	candidate := s.existingTvChannel()
	candidate.ID = ""

	ctx := context.TODO()
	srv := tvchannels.NewService(s.store)
	_, err := srv.ReplaceChannel(ctx, candidate)

	s.EqualError(err, tvchannels.ErrInvalidID.Error())
}

func (s *ServiceTestSuite) TestReplaceChannelWithNonExistinID() {
	candidate := s.existingTvChannel()

	ctx := context.TODO()
	existing := s.existingTvChannel()
	s.store.EXPECT().GetByID(ctx, existing.ID).Return(tvchannels.TvChannel{}, nil)
	srv := tvchannels.NewService(s.store)
	_, err := srv.ReplaceChannel(ctx, candidate)

	s.EqualError(err, tvchannels.ErrInvalidID.Error())
}

func (s *ServiceTestSuite) TestReplaceChannelWithInvalidChannel() {
	candidate := s.existingTvChannel()
	candidate.Name = ""

	ctx := context.TODO()
	existing := s.existingTvChannel()
	s.store.EXPECT().GetByID(ctx, existing.ID).Return(existing, nil)
	srv := tvchannels.NewService(s.store)
	_, err := srv.ReplaceChannel(ctx, candidate)

	s.EqualError(err, tvchannels.ErrInvalidName.Error())
}

func (s *ServiceTestSuite) TestReplaceChannelWithStoreGatewayError() {
	candidate := s.existingTvChannel()

	ctx := context.TODO()
	existing := s.existingTvChannel()
	s.store.EXPECT().GetByID(ctx, existing.ID).Return(existing, nil)
	s.store.EXPECT().Replace(ctx, gomock.Any()).Return(errors.New("Oops something went wrong!!!"))
	srv := tvchannels.NewService(s.store)
	_, err := srv.ReplaceChannel(ctx, candidate)

	s.EqualError(err, tvchannels.ErrInternal.Error())
}

func (s *ServiceTestSuite) TestReplaceChannelSuccessfully() {
	candidate := s.existingTvChannel()
	candidate.Name = "Telefe Online"

	ctx := context.TODO()
	existing := s.existingTvChannel()
	s.store.EXPECT().GetByID(ctx, existing.ID).Return(existing, nil)
	s.store.EXPECT().Replace(ctx, gomock.Any()).Return(nil)
	srv := tvchannels.NewService(s.store)
	replaced, _ := srv.ReplaceChannel(ctx, candidate)

	s.Equal(candidate.Name, replaced.Name)
	s.Equal(candidate.ListPosition, replaced.ListPosition)
	s.Equal(candidate.Status, replaced.Status)
	s.Equal(candidate.Signals, replaced.Signals)
	s.NotEqual(candidate.UpdatedAt, replaced.UpdatedAt)
	s.Equal(existing.CreatedAt, replaced.CreatedAt)
}

func (s *ServiceTestSuite) TestDeleteChannelWithEmptyId() {
	existing := s.existingTvChannel()
	existing.ID = ""

	ctx := context.TODO()
	srv := tvchannels.NewService(s.store)
	_, err := srv.DeleteChannel(ctx, existing.ID)

	s.EqualError(err, tvchannels.ErrInvalidID.Error())
}

func (s *ServiceTestSuite) TestDeleteChannelWithNonExistingID() {
	existing := s.existingTvChannel()

	ctx := context.TODO()
	srv := tvchannels.NewService(s.store)
	s.store.EXPECT().GetByID(ctx, existing.ID).Return(tvchannels.TvChannel{}, nil)
	_, err := srv.DeleteChannel(ctx, existing.ID)

	s.EqualError(err, tvchannels.ErrInvalidID.Error())
}

func (s *ServiceTestSuite) TestDeleteChannelWithStoreGatewayError() {
	existing := s.existingTvChannel()

	ctx := context.TODO()
	srv := tvchannels.NewService(s.store)
	s.store.EXPECT().GetByID(ctx, existing.ID).Return(existing, nil)
	s.store.EXPECT().Delete(ctx, existing.ID).Return(errors.New("Oops something went wrong!!!"))
	_, err := srv.DeleteChannel(ctx, existing.ID)

	s.EqualError(err, tvchannels.ErrInternal.Error())
}

func (s *ServiceTestSuite) TestDeleteChannelSuccessfully() {
	existing := s.existingTvChannel()

	ctx := context.TODO()
	srv := tvchannels.NewService(s.store)
	s.store.EXPECT().GetByID(ctx, existing.ID).Return(existing, nil)
	s.store.EXPECT().Delete(ctx, existing.ID).Return(nil)
	deleted, _ := srv.DeleteChannel(ctx, existing.ID)

	s.Equal(existing, deleted)
}

func (s *ServiceTestSuite) TestListChannelsWithStoreGatewayError() {
	ctx := context.TODO()
	srv := tvchannels.NewService(s.store)
	s.store.EXPECT().List(ctx).Return(nil, errors.New("Oops something went wrong!!!"))
	_, err := srv.ListChannels(ctx)

	s.EqualError(err, tvchannels.ErrInternal.Error())
}

func (s *ServiceTestSuite) TestListChannelsSuccessfully() {
	existingList := []tvchannels.TvChannel{s.existingTvChannel()}

	ctx := context.TODO()
	srv := tvchannels.NewService(s.store)
	s.store.EXPECT().List(ctx).Return(existingList, nil)
	list, _ := srv.ListChannels(ctx)

	s.Equal(existingList, list)
}

func TestService(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
