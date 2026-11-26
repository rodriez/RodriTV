package tvchannels_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rodriez/RodriTV/internal/tvchannels"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
)

type HandlerTestSuite struct {
	suite.Suite

	service *MockTvChannelServiceGateway
}

func (s *HandlerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.service = NewMockTvChannelServiceGateway(ctrl)
}

func (s *HandlerTestSuite) validAddChannelRequest() tvchannels.TvChannelRequest {
	return tvchannels.TvChannelRequest{
		Name:         "Telefe",
		ListPosition: 1,
		Signals: []struct {
			URL                string `json:"url,omitempty"`
			Priority           int    `json:"priority,omitempty"`
			NeedBrowser        bool   `json:"needBrowser,omitempty"`
			NeedRefresh        bool   `json:"needRefresh,omitempty"`
			RefreshRateSeconds int    `json:"refreshRateSeconds,omitempty"`
			BufferEnabled      bool   `json:"bufferEnabled,omitempty"`
			BufferSize         int    `json:"bufferSize,omitempty"`
		}{
			{
				URL:                "https://www.mitelefe.com/telefe-en-vivo",
				Priority:           1,
				NeedBrowser:        false,
				NeedRefresh:        true,
				RefreshRateSeconds: 20,
			},
			{
				URL:           "http://45.134.141.161:2200/ARG/TELEFE_HD/index.m3u8",
				Priority:      2,
				NeedBrowser:   false,
				BufferEnabled: true,
				BufferSize:    20,
			},
		},
	}
}

func (s *HandlerTestSuite) existingTvChannel() tvchannels.TvChannel {
	return tvchannels.TvChannel{
		ID:           "3b871028-be80-4257-8759-53d19f7fbd92",
		Name:         "Telefe",
		ListPosition: 1,
		Status:       tvchannels.TVChannelStatusActive,
		Signals: tvchannels.TvSignalSet{
			{
				URL:                "https://www.mitelefe.com/telefe-en-vivo",
				Priority:           1,
				NeedBrowser:        false,
				NeedRefresh:        true,
				RefreshRateSeconds: 20,
			},
			{
				URL:           "http://45.134.141.161:2200/ARG/TELEFE_HD/index.m3u8",
				Priority:      2,
				NeedBrowser:   false,
				BufferEnabled: true,
				BufferSize:    20,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (s *HandlerTestSuite) TestAddChannelWithInvalidJson() {
	body := strings.NewReader(`{id:"asdad"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", body)

	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.AddChannel(recorder, req)

	var apiErr tvchannels.ApiError
	json.NewDecoder(recorder.Body).Decode(&apiErr)

	s.Equal(http.StatusBadRequest, recorder.Code)
	s.Equal(http.StatusText(http.StatusBadRequest), apiErr.Title)
	s.Equal("invalid or malformed json", apiErr.Detail)
	s.NotEmpty(apiErr.Errors)
}

func (s *HandlerTestSuite) TestAddChannelWithServiceInternalError() {
	body, _ := json.Marshal(s.validAddChannelRequest())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", strings.NewReader(string(body)))

	s.service.EXPECT().AddChannel(req.Context(), gomock.Any()).Return(tvchannels.TvChannel{}, tvchannels.ErrInternal)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.AddChannel(recorder, req)

	var response tvchannels.ApiError
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusInternalServerError, recorder.Code)
	s.Equal(http.StatusText(http.StatusInternalServerError), response.Title)
	s.Equal("Internal error", response.Detail)
	s.Empty(response.Errors)
}

func (s *HandlerTestSuite) TestAddChannelWithServiceValidationError() {
	body, _ := json.Marshal(s.validAddChannelRequest())
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", strings.NewReader(string(body)))

	s.service.EXPECT().AddChannel(req.Context(), gomock.Any()).Return(tvchannels.TvChannel{}, tvchannels.ErrInvalidName)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.AddChannel(recorder, req)

	var response tvchannels.ApiError
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusBadRequest, recorder.Code)
	s.Equal(http.StatusText(http.StatusBadRequest), response.Title)
	s.Equal("One or more field are invalid", response.Detail)
	s.Contains(response.Errors, tvchannels.ErrInvalidName.Error())
}

func (s *HandlerTestSuite) TestAddChannelWithServiceSuccessfully() {
	addChannelReq := s.validAddChannelRequest()
	body, _ := json.Marshal(addChannelReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/channels", strings.NewReader(string(body)))

	createdChannel := addChannelReq.ToChannel()
	createdChannel.ID = "3b871028-be80-4257-8759-53d19f7fbd92"
	createdChannel.Status = tvchannels.TVChannelStatusActive
	createdChannel.CreatedAt = time.Now()
	createdChannel.UpdatedAt = time.Now()

	s.service.EXPECT().AddChannel(req.Context(), gomock.Any()).Return(createdChannel, nil)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.AddChannel(recorder, req)

	var response tvchannels.TvChannelResponse
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusCreated, recorder.Code)
	s.Equal(createdChannel.ID, response.ID)
	s.Equal(createdChannel.Name, response.Name)
	s.Equal(createdChannel.ListPosition, response.ListPosition)
	s.Equal(response.Status, tvchannels.TVChannelStatusActive)
	s.NotEmpty(createdChannel.Signals)
	s.Equal(len(addChannelReq.Signals), len(response.Signals))
	for i, sig := range addChannelReq.Signals {
		s.Equal(sig.URL, response.Signals[i].URL)
		s.Equal(sig.Priority, response.Signals[i].Priority)
		s.Equal(sig.NeedBrowser, response.Signals[i].NeedBrowser)
	}
	s.Equal(createdChannel.CreatedAt.UTC().Format(time.RFC3339Nano), response.CreatedAt.UTC().Format(time.RFC3339Nano))
	s.Equal(createdChannel.UpdatedAt.UTC().Format(time.RFC3339Nano), response.UpdatedAt.UTC().Format(time.RFC3339Nano))
}

func (s *HandlerTestSuite) TestGetChannelByIDWithServiceInternalError() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/{channel_id}", nil)
	req.SetPathValue("channel_id", "3b871028-be80-4257-8759-53d19f7fbd92")

	s.service.EXPECT().GetChannel(req.Context(), "3b871028-be80-4257-8759-53d19f7fbd92").Return(tvchannels.TvChannel{}, tvchannels.ErrInternal)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.GetChannelByID(recorder, req)

	var response tvchannels.ApiError
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusInternalServerError, recorder.Code)
	s.Equal(http.StatusText(http.StatusInternalServerError), response.Title)
	s.Equal("Internal error", response.Detail)
	s.Empty(response.Errors)
}

func (s *HandlerTestSuite) TestGetChannelByIDWithNonRegisteredID() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/{channel_id}", nil)
	req.SetPathValue("channel_id", "3b871028-be80-4257-8759-53d19f7fbd92")

	s.service.EXPECT().GetChannel(req.Context(), "3b871028-be80-4257-8759-53d19f7fbd92").Return(tvchannels.TvChannel{}, tvchannels.ErrInvalidID)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.GetChannelByID(recorder, req)

	var response tvchannels.ApiError
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusNotFound, recorder.Code)
	s.Equal(http.StatusText(http.StatusNotFound), response.Title)
	s.Equal("Resource not exist", response.Detail)
	s.NotEmpty(response.Errors)
}

func (s *HandlerTestSuite) TestGetChannelByIDWithServiceSuccessfully() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/{channel_id}", nil)
	req.SetPathValue("channel_id", "3b871028-be80-4257-8759-53d19f7fbd92")

	existing := s.existingTvChannel()

	s.service.EXPECT().GetChannel(req.Context(), "3b871028-be80-4257-8759-53d19f7fbd92").Return(existing, nil)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.GetChannelByID(recorder, req)

	var response tvchannels.TvChannelResponse
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusOK, recorder.Code)
	s.Equal(existing.ID, response.ID)
	s.Equal(existing.Name, response.Name)
	s.Equal(existing.ListPosition, response.ListPosition)
	s.Equal(existing.Status, response.Status)
	s.NotEmpty(existing.Signals)
	for i, sig := range response.Signals {
		s.Equal(sig.URL, response.Signals[i].URL)
		s.Equal(sig.Priority, response.Signals[i].Priority)
		s.Equal(sig.NeedBrowser, response.Signals[i].NeedBrowser)
	}
	s.Equal(existing.CreatedAt.UTC().Format(time.RFC3339Nano), response.CreatedAt.UTC().Format(time.RFC3339Nano))
	s.Equal(existing.UpdatedAt.UTC().Format(time.RFC3339Nano), response.UpdatedAt.UTC().Format(time.RFC3339Nano))
}

func (s *HandlerTestSuite) TestReplaceChannelWithInvalidJson() {
	body := strings.NewReader(`{name:"asdad"}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/channels/{channel_id}", body)
	req.SetPathValue("channel_id", "3b871028-be80-4257-8759-53d19f7fbd92")

	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.ReplaceChannel(recorder, req)

	var apiErr tvchannels.ApiError
	json.NewDecoder(recorder.Body).Decode(&apiErr)

	s.Equal(http.StatusBadRequest, recorder.Code)
	s.Equal(http.StatusText(http.StatusBadRequest), apiErr.Title)
	s.Equal("invalid or malformed json", apiErr.Detail)
	s.NotEmpty(apiErr.Errors)
}

func (s *HandlerTestSuite) TestReplaceChannelWithServiceInternalError() {
	body, _ := json.Marshal(s.validAddChannelRequest())
	req := httptest.NewRequest(http.MethodPut, "/api/v1/channels/{channel_id}", strings.NewReader(string(body)))

	s.service.EXPECT().ReplaceChannel(req.Context(), gomock.Any()).Return(tvchannels.TvChannel{}, tvchannels.ErrInternal)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.ReplaceChannel(recorder, req)

	var response tvchannels.ApiError
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusInternalServerError, recorder.Code)
	s.Equal(http.StatusText(http.StatusInternalServerError), response.Title)
	s.Equal("Internal error", response.Detail)
	s.Empty(response.Errors)
}

func (s *HandlerTestSuite) TestReplaceChannelWithServiceSuccessfully() {
	request := s.validAddChannelRequest()
	request.Status = tvchannels.TVChannelStatusActive
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/channels/{channel_id}", strings.NewReader(string(body)))
	req.SetPathValue("channel_id", "3b871028-be80-4257-8759-53d19f7fbd92")

	replacedChannel := request.ToChannel()
	replacedChannel.ID = "3b871028-be80-4257-8759-53d19f7fbd92"
	replacedChannel.Status = tvchannels.TVChannelStatusActive
	replacedChannel.CreatedAt = time.Now()
	replacedChannel.UpdatedAt = time.Now()

	s.service.EXPECT().ReplaceChannel(req.Context(), gomock.Any()).Return(replacedChannel, nil)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.ReplaceChannel(recorder, req)

	var response tvchannels.TvChannelResponse
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusOK, recorder.Code)
	s.Equal(replacedChannel.ID, response.ID)
	s.Equal(replacedChannel.Name, response.Name)
	s.Equal(replacedChannel.ListPosition, response.ListPosition)
	s.Equal(request.Status, response.Status)
	s.NotEmpty(replacedChannel.Signals)
	s.Equal(len(request.Signals), len(response.Signals))
	for i, sig := range request.Signals {
		s.Equal(sig.URL, response.Signals[i].URL)
		s.Equal(sig.Priority, response.Signals[i].Priority)
		s.Equal(sig.NeedBrowser, response.Signals[i].NeedBrowser)
	}
	s.Equal(replacedChannel.CreatedAt.UTC().Format(time.RFC3339Nano), response.CreatedAt.UTC().Format(time.RFC3339Nano))
	s.Equal(replacedChannel.UpdatedAt.UTC().Format(time.RFC3339Nano), response.UpdatedAt.UTC().Format(time.RFC3339Nano))
}

func (s *HandlerTestSuite) TestDeleteChannelWithServiceInternalError() {
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/channels/{channel_id}", nil)
	req.SetPathValue("channel_id", "3b871028-be80-4257-8759-53d19f7fbd92")

	s.service.EXPECT().DeleteChannel(req.Context(), "3b871028-be80-4257-8759-53d19f7fbd92").Return(tvchannels.TvChannel{}, tvchannels.ErrInternal)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.DeleteChannel(recorder, req)

	var response tvchannels.ApiError
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusInternalServerError, recorder.Code)
	s.Equal(http.StatusText(http.StatusInternalServerError), response.Title)
	s.Equal("Internal error", response.Detail)
	s.Empty(response.Errors)
}

func (s *HandlerTestSuite) TestDeleteChannelWithServiceSuccessfully() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels/{channel_id}", nil)
	req.SetPathValue("channel_id", "3b871028-be80-4257-8759-53d19f7fbd92")

	existing := s.existingTvChannel()

	s.service.EXPECT().DeleteChannel(req.Context(), "3b871028-be80-4257-8759-53d19f7fbd92").Return(existing, nil)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.DeleteChannel(recorder, req)

	var response tvchannels.TvChannelResponse
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusOK, recorder.Code)
	s.Equal(existing.ID, response.ID)
	s.Equal(existing.Name, response.Name)
	s.Equal(existing.ListPosition, response.ListPosition)
	s.Equal(existing.Status, response.Status)
	s.NotEmpty(existing.Signals)
	for i, sig := range response.Signals {
		s.Equal(sig.URL, response.Signals[i].URL)
		s.Equal(sig.Priority, response.Signals[i].Priority)
		s.Equal(sig.NeedBrowser, response.Signals[i].NeedBrowser)
	}
	s.Equal(existing.CreatedAt.UTC().Format(time.RFC3339Nano), response.CreatedAt.UTC().Format(time.RFC3339Nano))
	s.Equal(existing.UpdatedAt.UTC().Format(time.RFC3339Nano), response.UpdatedAt.UTC().Format(time.RFC3339Nano))
}

func (s *HandlerTestSuite) TestListChannelsWithServiceInternalError() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)

	s.service.EXPECT().ListChannels(req.Context()).Return(nil, tvchannels.ErrInternal)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.ListChannels(recorder, req)

	var response tvchannels.ApiError
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusInternalServerError, recorder.Code)
	s.Equal(http.StatusText(http.StatusInternalServerError), response.Title)
	s.Equal("Internal error", response.Detail)
	s.Empty(response.Errors)
}

func (s *HandlerTestSuite) TestListChannelsWithServiceSuccessfully() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/channels", nil)

	existing := s.existingTvChannel()

	s.service.EXPECT().ListChannels(req.Context()).Return([]tvchannels.TvChannel{existing}, nil)
	recorder := httptest.NewRecorder()
	handler := tvchannels.NewHandler(s.service)
	handler.ListChannels(recorder, req)

	var response []tvchannels.TvChannelResponse
	json.NewDecoder(recorder.Body).Decode(&response)

	s.Equal(http.StatusOK, recorder.Code)
	s.Equal(1, len(response))
	s.Equal(existing.ID, response[0].ID)
	s.Equal(existing.Name, response[0].Name)
	s.Equal(existing.ListPosition, response[0].ListPosition)
	s.Equal(existing.Status, response[0].Status)
	s.NotEmpty(existing.Signals)
	for i, sig := range response[0].Signals {
		s.Equal(sig.URL, response[0].Signals[i].URL)
		s.Equal(sig.Priority, response[0].Signals[i].Priority)
		s.Equal(sig.NeedBrowser, response[0].Signals[i].NeedBrowser)
	}
	s.Equal(existing.CreatedAt.UTC().Format(time.RFC3339Nano), response[0].CreatedAt.UTC().Format(time.RFC3339Nano))
	s.Equal(existing.UpdatedAt.UTC().Format(time.RFC3339Nano), response[0].UpdatedAt.UTC().Format(time.RFC3339Nano))
}

func TestHandler(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
