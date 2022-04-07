package routes

import (
	"github.com/RishabhBajpai97/golang-jwt-authentication/controller"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(incomingRoutes *gin.Engine){
	incomingRoutes.POST("/users/signup",controller.Signup())
	incomingRoutes.POST("/users/login",controller.Login())
}