package study_material_repo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	study_material_models "github.com/Petr09Mitin/xrust-beze-back/internal/models/study_material"
	"github.com/Petr09Mitin/xrust-beze-back/internal/pkg/config"
	"github.com/rs/zerolog"
	"io"
	"net/http"
)

type MLTaggerRepo interface {
	ProcessAttachment(ctx context.Context, attachment *study_material_models.AttachmentToParse) (*study_material_models.ParsedAttachmentResponse, error)
}

type MLTaggerRepoImpl struct {
	httpClient *http.Client
	cfg        *config.GRPCService
	logger     zerolog.Logger
}

func NewMLTaggerRepo(cfg *config.GRPCService, logger zerolog.Logger) MLTaggerRepo {
	return &MLTaggerRepoImpl{
		httpClient: http.DefaultClient,
		cfg:        cfg,
		logger:     logger,
	}
}

func (m *MLTaggerRepoImpl) ProcessAttachment(ctx context.Context, attachment *study_material_models.AttachmentToParse) (*study_material_models.ParsedAttachmentResponse, error) {
	marshalled, err := json.Marshal(study_material_models.AttachmentToParseRequest{
		Filename: attachment.Filename,
		S3Bucket: config.AttachmentsMinioBucket,
	})
	if err != nil {
		m.logger.Error().Err(err).Msg("unable to marshal json to parse study material")
		return nil, err
	}
	url := fmt.Sprintf("http://%s:%d/set-tag", m.cfg.Host, m.cfg.Port)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(marshalled))
	if err != nil {
		m.logger.Error().Err(err).Msg("unable to build parse study material req")
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := m.httpClient.Do(req)
	if err != nil {
		m.logger.Error().Err(err).Msg("unable to parse study material")
		return nil, err
	}
	defer func() {
		err = res.Body.Close()
		if err != nil {
			m.logger.Error().Err(err).Msg("unable to close res body in parse study material")
		}
	}()
	if res.StatusCode >= http.StatusBadRequest {
		err = fmt.Errorf("%d: %s", res.StatusCode, res.Status)
		m.logger.Error().Err(err).Msg("unable to parse study material")
		return nil, custom_errors.ErrParsingStudyMaterialsUnavailable
	}
	parsedAttach := &study_material_models.ParsedAttachmentResponse{}
	bytesRes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytesRes, parsedAttach)
	if err != nil {
		return nil, err
	}
	if parsedAttach.StudyMaterial != nil {
		// tags and name should be inferred from model
		parsedAttach.StudyMaterial.Filename = attachment.Filename
		parsedAttach.StudyMaterial.AuthorID = attachment.AuthorID
	}
	return parsedAttach, nil
}
