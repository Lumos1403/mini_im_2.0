package response

import (
	"net/http"

	apperrors "mini_im/backend/internal/pkg/errors"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(ctx *gin.Context, data interface{}) {
	ctx.JSON(http.StatusOK, Response{
		Code:    apperrors.CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func Fail(ctx *gin.Context, status int, err *apperrors.AppError) {
	if err == nil {
		err = apperrors.ErrInternal
	}

	ctx.JSON(status, Response{
		Code:    err.Code,
		Message: err.Message,
		Data:    nil,
	})
}
