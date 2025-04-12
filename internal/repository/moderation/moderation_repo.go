package ml_moderation_repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	user_model "github.com/Petr09Mitin/xrust-beze-back/internal/models/user"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/rs/zerolog"
)

type ModerationRepo interface {
	CheckSwearing(ctx context.Context, text string) (bool, error)
}

type moderationRepository struct {
	httpClient *http.Client
	cfg        *config.GRPCService
	logger     zerolog.Logger
}

func NewModerationRepository(cfg *config.GRPCService, logger zerolog.Logger) ModerationRepo {
	return &moderationRepository{
		httpClient: http.DefaultClient,
		cfg:        cfg,
		logger:     logger,
	}
}

func (m *moderationRepository) CheckSwearing(ctx context.Context, text string) (bool, error) {
	reqBody := user_model.CheckSwearingRequest{
		Text: text,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		m.logger.Error().Err(err).Msg("unable to marshal json to check swearing")
		return false, err
	}

	url := fmt.Sprintf("http://%s:%d/check", m.cfg.Host, m.cfg.Port)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		m.logger.Error().Err(err).Msg("unable to build ml_moderation req")
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := m.httpClient.Do(req)
	if err != nil {
		m.logger.Error().Err(err).Msg("unable to check swearing")
		return false, err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			m.logger.Error().Err(err).Msg("unable to close res body")
		}
	}()

	if res.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("%d: %s", res.StatusCode, res.Status)
		m.logger.Error().Err(err).Msg("unable to check swearing")
		return false, custom_errors.ErrModerationUnavailable
	}

	var response user_model.CheckSwearingResponse
	bytesRes, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	err = json.Unmarshal(bytesRes, &response)
	if err != nil {
		return false, err
	}

	return response.IsProfanity, nil
	// return true, nil
}
