package rag_client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/rs/zerolog"
	"net/http"
)

type RagRepo interface {
	NotifyStudyMaterialAdded(ctx context.Context, data *study_material_models.NewMaterialRAGData) (err error)
	NotifyStudyMaterialDeleted(ctx context.Context, data *study_material_models.DeleteMaterialRAGData) (err error)
}

type RagRepoImpl struct {
	httpClient *http.Client
	cfg        *config.GRPCService
	logger     zerolog.Logger
}

func NewRagClient(cfg *config.GRPCService, logger zerolog.Logger) RagRepo {
	return &RagRepoImpl{
		httpClient: http.DefaultClient,
		cfg:        cfg,
		logger:     logger,
	}
}

func (c *RagRepoImpl) NotifyStudyMaterialAdded(ctx context.Context, data *study_material_models.NewMaterialRAGData) (err error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		c.logger.Error().Err(err).Msg("unable to marshal json to notify study_material_added")
		return err
	}

	url := fmt.Sprintf("http://%s:%d/add", c.cfg.Host, c.cfg.Port)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		c.logger.Error().Err(err).Msg("unable to build study_material_added req")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Msg("unable to notify study_material_added http.Do")
		return err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			c.logger.Error().Err(err).Msg("unable to close res body")
		}
	}()

	if res.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("%d: %s", res.StatusCode, res.Status)
		c.logger.Error().Err(err).Msg("unable to notify study_material_added")
		return custom_errors.ErrRAGUnavailable
	}

	return nil
}

func (c *RagRepoImpl) NotifyStudyMaterialDeleted(ctx context.Context, data *study_material_models.DeleteMaterialRAGData) (err error) {
	jsonBody, err := json.Marshal(data)
	if err != nil {
		c.logger.Error().Err(err).Msg("unable to marshal json to notify study_material_deleted")
		return err
	}

	url := fmt.Sprintf("http://%s:%d/delete", c.cfg.Host, c.cfg.Port)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		c.logger.Error().Err(err).Msg("unable to build study_material_deleted req")
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Msg("unable to notify study_material_deleted http.Do")
		return err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			c.logger.Error().Err(err).Msg("unable to close res body")
		}
	}()

	if res.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("%d: %s", res.StatusCode, res.Status)
		c.logger.Error().Err(err).Msg("unable to notify study_material_deleted")
		return custom_errors.ErrRAGUnavailable
	}

	return nil
}
