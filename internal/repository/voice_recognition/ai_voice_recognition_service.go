package voice_recognition_repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	chat_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/chat"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"io"
	"net/http"

	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/rs/zerolog"
)

type VoiceRecognitionRepo interface {
	SendVoiceRecognitionRequest(ctx context.Context, filename string) (string, error)
}

type VoiceRecognitionRepoImpl struct {
	httpClient *http.Client
	cfg        *config.GRPCService
	logger     zerolog.Logger
}

func NewVoiceRecognitionRepo(cfg *config.GRPCService, logger zerolog.Logger) VoiceRecognitionRepo {
	return &VoiceRecognitionRepoImpl{
		httpClient: http.DefaultClient,
		cfg:        cfg,
		logger:     logger,
	}
}

func (s *VoiceRecognitionRepoImpl) SendVoiceRecognitionRequest(ctx context.Context, filename string) (string, error) {
	marshalled, err := json.Marshal(chat_models.VoiceRecognitionRequest{
		Bucket:   config.VoiceMessagesMinioBucket,
		Filename: filename,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("unable to marshal json to recognize voice")
		return "", err
	}
	url := fmt.Sprintf("http://%s:%d/transcribe/", s.cfg.Host, s.cfg.Port)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(marshalled))
	if err != nil {
		s.logger.Error().Err(err).Msg("unable to build voice recognition req")
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error().Err(err).Msg("unable to do voice recognition req")
		return "", err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			s.logger.Error().Err(err).Msg("unable to close res body")
		}
	}()
	if res.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("%d: %s", res.StatusCode, res.Status)
		s.logger.Error().Err(err).Msg("unable to do voice recognition req")
		return "", custom_errors.ErrStructurizationUnavailable
	}
	structMsg := &chat_models.VoiceRecognitionResponse{}
	bytesRes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(bytesRes, structMsg)
	if err != nil {
		return "", err
	}

	return structMsg.Text, nil
}
