package handlers

import (
	"errors"
	"net/http"
	"strings"

	"github.com/AdhityaRamadhanus/chronicle/server/render"
)

var (
	// request body processing error
	ErrFailedToUnmarshalJSON = errors.New("Failed to Unmarshal JSON")
	ErrFailedToReadBody      = errors.New("Failed to read request body")
	ErrSomethingWrong        = errors.New("Something is wrong")
	ErrInvalidRequest        = errors.New("Invalid Request Against this endpoint")
)

func RenderError(res http.ResponseWriter, err error, customMessages ...string) error {
	errorMessage := err.Error()
	if len(customMessages) > 0 {
		errorMessage = strings.Join(customMessages, " ")
	}

	switch err {
	case ErrFailedToReadBody:
		return render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
			"status": http.StatusInternalServerError,
			"error": map[string]interface{}{
				"code":    "ErrFailedToReadBody",
				"message": errorMessage,
			},
		})
	case ErrFailedToUnmarshalJSON:
		return render.JSON(res, http.StatusBadRequest, map[string]interface{}{
			"status": http.StatusBadRequest,
			"error": map[string]interface{}{
				"code":    "ErrFailedToUnmarshalJSON",
				"message": errorMessage,
			},
		})
	case ErrSomethingWrong:
		return render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
			"status": http.StatusInternalServerError,
			"error": map[string]interface{}{
				"code":    "ErrInternalServer",
				"message": errorMessage,
			},
		})
	case ErrInvalidRequest:
		return render.JSON(res, 422, map[string]interface{}{
			"status": 422,
			"error": map[string]interface{}{
				"code":    "ErrInvalidRequest",
				"message": errorMessage,
			},
		})
	default:
		return render.JSON(res, http.StatusInternalServerError, map[string]interface{}{
			"status": http.StatusInternalServerError,
			"error": map[string]interface{}{
				"code":    "ErrInternalServer",
				"message": errorMessage,
			},
		})
	}
}
