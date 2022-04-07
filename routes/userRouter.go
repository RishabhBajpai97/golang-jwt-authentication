package routes

import (
	"github.com/RishabhBajpai97/golang-jwt-authentication/controller"
	"github.com/RishabhBajpai97/golang-jwt-authentication/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {
		incomingRoutes.Use(middleware.Authenticate())
		incomingRoutes.GET("/users",controller.GetUsers())
		incomingRoutes.GET("/users/:user_id",controller.GetUserById())
}