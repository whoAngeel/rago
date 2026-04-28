package rest

import "github.com/gin-gonic/gin"

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details"`
}

func RespondError(c *gin.Context, status int, msg, details string) {
	c.JSON(status, ErrorResponse{Error: msg, Details: details})
}
