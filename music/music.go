package music

import (
	"beep/controllers"
	"beep/database"
	"beep/models"
	"beep/views"
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	vkCloudHotboxEndpoint = "https://hb.vkcs.cloud"
	defaultRegion         = "ru-msk"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var genreCollection *mongo.Collection = database.GenreData(database.Client, "Genre")
var keyCollection *mongo.Collection = database.KeyData(database.Client, "Key")
var instrumentCollection *mongo.Collection = database.InstrumenttData(database.Client, "Instrument")
var variationCollection *mongo.Collection = database.VariationtData(database.Client, "Ivariations")

func AudioPlay() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
func GetUploadPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получение данных из коллекций или других источников
		genres, err := controllers.GenreToSlice(genreCollection)
		if err != nil {
			log.Println("Failed to get genres:", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Internal Server Error"})
			return
		}
		variations, err := controllers.VariationsToSlice(variationCollection)
		if err != nil {
			log.Println("Failed to get variations:", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Internal Server Error"})
			return
		}
		userId := c.Query("id")
		keys, err := controllers.KeyToSlice(keyCollection)
		if err != nil {
			log.Println("Failed to get keys:", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Internal Server Error"})
			return
		}

		instruments, err := controllers.InstrumentToSlice(instrumentCollection)
		if err != nil {
			log.Println("Failed to get instruments:", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Internal Server Error"})
			return
		}
		views.Upload(genres, keys, instruments, variations, userId).Render(c, c.Writer)
		// Отправка данных в шаблон
		// c.HTML(http.StatusOK, "upload.html", gin.H{
		// 	"Genres":      genres,
		// 	"Keys":        keys,
		// 	"Instruments": instruments,
		// 	"userId":      userId,
		// 	"Variations":  variations,
		// })
	}
}

func UploadedFiles() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		err := c.Request.ParseMultipartForm(10 << 20) // разбор формы с максимальным размером 10 MB
		if err != nil {
			c.JSON(http.StatusInternalServerError, "error")
			fmt.Println("1")
			return
		}
		imageFile, imageHeader, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, "error")
			fmt.Println("картинка")
			return
		}
		defer imageFile.Close()

		audioFile, audioHeader, err := c.Request.FormFile("audio")
		if err != nil {
			c.JSON(http.StatusBadRequest, "error")
			fmt.Println("Звучок")
			return
		}
		defer audioFile.Close()

		imageURL := fmt.Sprintf("/temp/%s", imageHeader.Filename)
		audioURL := fmt.Sprintf("/temp/%s", audioHeader.Filename)
		err = uploadFileToVKStorage(imageURL, imageFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "error")
			fmt.Println("не загрузилась картинка")
			return
		}

		err = uploadFileToVKStorage(audioURL, audioFile)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "error")
			fmt.Println("не загрузился звучок")
			return
		}

		userID := c.Query("id")
		fmt.Println(userID)
		if userID == "" {
			c.JSON(http.StatusBadRequest, "user id is empty")
			fmt.Println("нет id")
			return
		}
		var selectedGenres = c.PostFormArray("genre[]")
		selectedKeys := c.PostForm("key")
		var selectedInstruments = c.PostFormArray("instrument[]")
		var selectedVariations = c.PostFormArray("variation[]")
		var newProduct models.Product
		newProduct.Product_ID = primitive.NewObjectID()
		newProduct.Author_ID, err = primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.IndentedJSON(500, "Internal server error")
			fmt.Println("не создался айди 1 ")

		}
		beatsPerMinute := c.PostForm("bpm")
		bpm, err := strconv.ParseUint(beatsPerMinute, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "что-то не так с bpm"})
			fmt.Println("что-то не так с bpm")
			return
		}
		priceStr := c.PostForm("price")
		price, err := strconv.ParseUint(priceStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "чето не так с ценой"})
			fmt.Println("чето не так с ценой")
			return
		}
		newProduct.Genre = &selectedGenres
		newProduct.Key = &selectedKeys
		newProduct.Instrument = &selectedInstruments
		newProduct.Variation = &selectedVariations
		newProduct.Bpm = aws.Uint64(bpm)
		newProduct.Price = aws.Uint64(price)
		newProduct.Product_Name = &audioHeader.Filename
		fileLink := fmt.Sprintf("https://razeim.hb.ru-msk.vkcs.cloud/%s", audioHeader.Filename)
		imageLink := fmt.Sprintf("https://razeim.hb.ru-msk.vkcs.cloud/%s", imageHeader.Filename)
		newProduct.File_Link = &fileLink
		newProduct.Image = &imageLink
		_, err = ProductCollection.InsertOne(ctx, newProduct)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "error")
			fmt.Println("продуктик не воткнулся")
			return
		}

		userObjectID, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(500, "Internal Server error")
			fmt.Println("не создался айди 2")
			return
		}
		var user models.User
		err = UserCollection.FindOne(ctx, bson.M{"_id": userObjectID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Failed to find user")
			fmt.Println("не нашелся пользователь ")
			return
		}
		user.Uploaded_Files = append(user.Uploaded_Files, newProduct.Product_ID)
		update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "uploaded_files", Value: user.Uploaded_Files}}}}
		_, err = UserCollection.UpdateOne(ctx, bson.M{"_id": userObjectID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, "Failed to update user")
			fmt.Println("не обновился")
			return
		}

		c.JSON(http.StatusOK, "все отлично")
	}
}
func uploadFileToVKStorage(fileURL string, uploadedFile multipart.File) error {
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(vkCloudHotboxEndpoint),
		Region:   aws.String(defaultRegion),
	})
	if err != nil {
		fmt.Println("или вот здесь")
		return err
	}

	svc := s3.New(sess)
	bucket := "razeim"

	// Определение имени файла на основе URL
	fileName := filepath.Base(fileURL)

	// Загрузка файла в бакет
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   uploadedFile,              // Используем загруженный файл
		ACL:    aws.String("public-read"), // Установка ACL на public-read
	})
	if err != nil {
		fmt.Println("проблемка здесь")
		return err
	}

	log.Printf("File %s uploaded successfully to bucket %q", fileName, bucket)
	return nil
}
func AddProduct(c *gin.Context, userID string) {

}
