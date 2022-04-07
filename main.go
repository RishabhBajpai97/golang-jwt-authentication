package main

import (
	"os"
	"github.com/RishabhBajpai97/golang-jwt-authentication/routes"
	"github.com/gin-gonic/gin"
)
func main()  {
	port:=os.Getenv("PORt")
	if port==""{
		port= "8000"
	}
	router:= gin.New()
	router.Use(gin.Logger())
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.GET("/api-1",func(ctx *gin.Context) {
		ctx.JSON(200,gin.H{"success":"Access granted to api-1"})
	})
	router.GET("/api-2",func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"Success":"Access granted to api-2"})
	})

	router.Run(":"+port)

}