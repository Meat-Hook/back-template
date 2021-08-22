package web

import (
	"errors"
	"net/http"

	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/models"
	"github.com/Meat-Hook/back-template/cmd/file/internal/api/web/generated/restapi/operations"
	"github.com/Meat-Hook/back-template/cmd/file/internal/app"
	"github.com/go-openapi/swag"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
)

func (svc *service) getFile(params operations.GetFileParams) operations.GetFileResponder {
	ctx, log, _ := fromRequest(params.HTTPRequest)

	fID, err := uuid.FromString(params.ID.String())
	if err != nil {
		return operations.NewGetFileDefault(http.StatusBadRequest).WithPayload(apiError(app.ErrNotValidID.Error()))
	}

	file, err := svc.app.GetFile(ctx, fID)
	defer logs(log, err)
	switch {
	case err == nil:
		return operations.NewGetFileOK().WithPayload(file)
	case errors.Is(err, app.ErrNotFound):
		return operations.NewGetFileDefault(http.StatusNotFound).WithPayload(apiError(app.ErrNotFound.Error()))
	default:
		return operations.NewGetFileDefault(http.StatusInternalServerError).
			WithPayload(apiError(http.StatusText(http.StatusInternalServerError)))
	}
}

func apiError(txt string) *models.Error {
	return &models.Error{
		Message: swag.String(txt),
	}
}

func logs(log zerolog.Logger, err error) {
	if err != nil {
		log.Error().Err(err).Send()
	}
}
