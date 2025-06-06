package filehandler

import (
	custom_errors "github.com/Petr09Mitin/xrust-beze-back/internal/models/error"
	"github.com/Petr09Mitin/xrust-beze-back/internal/router/middleware"
	"github.com/Petr09Mitin/xrust-beze-back/internal/services/file"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"net/http"
	"os"
)

type FileHandler struct {
	fileService file.FileService
	logger      zerolog.Logger
}

func NewFileHandler(router *gin.Engine, fileService file.FileService, logger zerolog.Logger) {
	h := &FileHandler{
		fileService: fileService,
		logger:      logger,
	}

	router.Use(middleware.CORSMiddleware())
	gr := router.Group("/api/v1/file")
	{
		gr.POST("/temp", h.UploadTempFile)
	}
}

func (f *FileHandler) UploadTempFile(c *gin.Context) {
	ff, err := c.FormFile("file")
	if err != nil {
		f.logger.Error().Err(err).Msg("form file error")
		custom_errors.WriteHTTPError(c, err)
		return
	}

	filepath := "/app/tempfiles/" + uuid.New().String() + ff.Filename
	err = c.SaveUploadedFile(ff, filepath)
	if err != nil {
		f.logger.Error().Err(err).Msg("form file error")
		custom_errors.WriteHTTPError(c, err)
		return
	}
	defer func() {
		err = os.Remove(filepath)
		if err != nil {
			f.logger.Error().Err(err).Msg("remove temp file error")
			return
		}
	}()

	filename, err := f.fileService.UploadTempFile(c.Request.Context(), filepath)
	if err != nil {
		f.logger.Error().Err(err).Msg("upload file error")
		custom_errors.WriteHTTPError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"filename": filename,
	})
}
