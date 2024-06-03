package main

import (
	"beep/controllers"
	"beep/database"
	"beep/middleware"
	"beep/music"
	"beep/routes"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
)

const vkCloudHotboxEndpoint = "https://hb.ru-msk.vkcs.cloud"
const defaultRegion = "us-east-1"

func main() {
	sess, _ := session.NewSession()
	svc := s3.New(sess, aws.NewConfig().WithEndpoint(vkCloudHotboxEndpoint).WithRegion(defaultRegion))
	bucket := "razeim"
	result, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Fatalf("Unable to list items in bucket %q, %v", bucket, err)
	} else {
		// итерирование по объектам
		for _, item := range result.Contents {
			log.Printf("Object: %s, size: %d\n", aws.StringValue(item.Key), aws.Int64Value(item.Size))
		}
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"

	}
	app := controllers.NewApplication(
		database.ProductData(database.Client, "Products"),
		database.UserData(database.Client, "Users"),
		database.KeyData(database.Client, "Key"),
		database.GenreData(database.Client, "Genre"),
		database.InstrumenttData(database.Client, "Instrument"),
		database.VariationtData(database.Client, "Ivariations"),
	)

	router := gin.New()
	router.Use(gin.Logger())
	routes.UserRoutes(router)
	router.Use(middleware.Authentication())
	router.LoadHTMLFiles("./smth/audio.html", "./smth/upload.html", "./smth/home.html", "./smth/success.html", "./smth/login.html", "./smth/signup.html")
	router.Static("/static", "./static")
	router.StaticFile("/audioscript.js", "./audioscript.js")
	router.GET("/upload", music.GetUploadPage())
	router.POST("/upload", music.UploadedFiles())
	router.GET("/sounds", controllers.GetSoundsPage())
	router.GET("/addtocart", app.AddToCart())
	router.GET("/removeitem", app.RemoteItem())
	router.GET("/listcart", controllers.GetItemFromCart())
	router.POST("/addpaymentdetails", controllers.AddPaymentDetails())
	router.PUT("/editpaymentdetails", controllers.EditPaymentDetails())
	router.GET("/deletpaymentdetails", controllers.DeletePaymentDetails())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())
	router.GET("/audio", music.AudioPlay())
	router.GET("/users/updatecart", controllers.UpdateTotalCart())
	router.GET("/users/profile", controllers.ProfilePage())
	router.DELETE("/users/profile", controllers.DeleteProduct())
	router.GET("/users/search", controllers.SearchProductByQuery())
	router.GET("/users/cart", controllers.GetCartPage())
	log.Fatal(router.Run(":" + port))
}
