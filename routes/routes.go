package routes

import (
	"beep/controllers"
	"beep/views"

	"github.com/gin-gonic/gin"
)

func UserRoutes(incomingRoutes *gin.Engine) {

	incomingRoutes.POST("/users/signup", controllers.SignUp())
	incomingRoutes.POST("/users/login", controllers.Login())
	incomingRoutes.POST("/admin/addproduct", controllers.ProductViewerAdmin())
	incomingRoutes.GET("/users/productview", controllers.SearchProduct())

	incomingRoutes.GET("/users/signup", func(c *gin.Context) {
		views.SignUp().Render(c, c.Writer)
	})
	incomingRoutes.GET("/users/login", func(c *gin.Context) {
		views.Login().Render(c, c.Writer)
	})
	incomingRoutes.GET("/", controllers.HomePage())
}
