package structurization_repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	structurizationmodels "github.com/Petr09Mitin/xrust-beze-back/internal/models/structurization"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/rs/zerolog"
)

type StructurizationRepository interface {
	SendStructRequest(ctx context.Context, question, answer string) (*structurizationmodels.StructurizedMessage, error)
}

type StructurizationRepositoryImpl struct {
	httpClient *http.Client
	cfg        *config.GRPCService
	logger     zerolog.Logger
}

func NewStructurizationRepository(cfg *config.GRPCService, logger zerolog.Logger) StructurizationRepository {
	return &StructurizationRepositoryImpl{
		httpClient: http.DefaultClient,
		cfg:        cfg,
		logger:     logger,
	}
}

func (s *StructurizationRepositoryImpl) SendStructRequest(ctx context.Context, question, answer string) (*structurizationmodels.StructurizedMessage, error) {
	marshalled, err := json.Marshal(structurizationmodels.StructRequest{
		Query:  question,
		Answer: answer,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("unable to marshal json to structurize")
		return nil, err
	}
	url := fmt.Sprintf("http://%s:%d/explane", s.cfg.Host, s.cfg.Port)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(marshalled))
	if err != nil {
		s.logger.Error().Err(err).Msg("unable to build structurization req")
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error().Err(err).Msg("unable to structurize")
		return nil, err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			s.logger.Error().Err(err).Msg("unable to close res body")
		}
	}()
	if res.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("%d: %s", res.StatusCode, res.Status)
		s.logger.Error().Err(err).Msg("unable to structurize")
		return nil, custom_errors.ErrStructurizationUnavailable
	}
	structMsg := &structurizationmodels.StructurizedMessage{}
	bytesRes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytesRes, structMsg)
	if err != nil {
		return nil, err
	}

	return structMsg, nil
}
