package controllers

import (
	"beep/components"
	"beep/database"
	"beep/models"
	generate "beep/tokens"
	"beep/views"
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "Users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "Products")
var genreCollection *mongo.Collection = database.GenreData(database.Client, "Genre")
var keyCollection *mongo.Collection = database.KeyData(database.Client, "Key")
var instrumentCollection *mongo.Collection = database.InstrumenttData(database.Client, "Instrument")
var Validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)

}

func VerifyPassword(userPassword string, givenPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	valid := true
	msg := ""
	if err != nil {
		msg = "пароль неправильный"
		valid = false
	}
	return valid, msg
}

func SignUp() gin.HandlerFunc {

	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var user models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := Validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": *user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})

		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": " this phone already in use "})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_At, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_ID = user.ID.Hex()
		token, refreshtoken, err := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, user.User_ID)
		if err != nil {
			log.Println(err)
		}
		user.Token = &token
		if token == "" {
			fmt.Println("пустой токен ")
		}
		user.Refresh_Token = &refreshtoken
		user.User_Cart = make([]models.Product_User, 0)
		user.Payment_Details = make([]models.PaymentDetails, 0)
		user.Order_Status = make([]models.Order, 0)
		user.Uploaded_Files = make([]primitive.ObjectID, 0)

		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user didnt benn created"})
			return
		}
		defer cancel()

		c.JSON(http.StatusCreated, "Successfuly signed in")
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var founduser models.User
		//var user models.User
		password := c.PostForm("password")
		email := c.PostForm("email")

		// if err := c.BindJSON(&user); err != nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{"error": err})
		// 	return
		// }
		err := UserCollection.FindOne(ctx, bson.M{"email": email}).Decode(&founduser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login or password incorrect"})
			return
		}

		PasswordIsValid, msg := VerifyPassword(password, *founduser.Password)

		defer cancel()
		fmt.Println(msg)
		if !PasswordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			return

		}

		token, refreshToken, _ := generate.TokenGenerator(*founduser.Email, *founduser.First_Name, *founduser.Last_Name, founduser.User_ID)
		defer cancel()

		generate.UpdateAllTokens(token, refreshToken, founduser.User_ID)
		c.Set("uid", founduser.ID)
		c.JSON(http.StatusOK, gin.H{"token": token})
	}
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var products models.Product
		defer cancel()
		if err := c.BindJSON(&products); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		products.Product_ID = primitive.NewObjectID()
		_, anyErr := ProductCollection.InsertOne(ctx, products)
		if anyErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "not inserted"})
			return
		}
		defer cancel()

		c.JSON(http.StatusOK, "successfully added")
	}
}

func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {

		var productList []models.Product
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := ProductCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "что-то пошло не так попробуйте позже еще раз")
			return
		}

		err = cursor.All(ctx, &productList)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		defer cursor.Close(ctx)

		if err := cursor.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(400, "недействительно")
			return
		}
		defer cancel()

		c.IndentedJSON(200, productList)

	}
}

func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {

		userIDRaw, _ := c.Get("uid")
		if userIDRaw == "" {
			fmt.Println("нет такого")
			return
		}
		userID, ok := userIDRaw.(string)
		if !ok {
			fmt.Println("userID не является строкой")
			return
		}

		var searchProducts []models.Product

		queryParam := c.Query("name")
		instrumentsParam := c.QueryArray("instruments")
		genresParam := c.QueryArray("genres")
		keyParam := c.Query("key")
		bpmParam := c.Query("bpm")

		filter := bson.M{}

		if queryParam != "" {
			filter["product_name"] = bson.M{
				"$regex":   queryParam,
				"$options": "i", // Нечувствительность к регистру
			}
		}
		if len(instrumentsParam) > 0 {
			filter["instrument"] = bson.M{"$in": instrumentsParam}
		}
		if len(genresParam) > 0 {
			filter["genre"] = bson.M{"$in": genresParam}
		}
		if keyParam != "" {
			filter["key"] = keyParam
		}
		if bpmParam != "" {
			bpm, err := strconv.Atoi(bpmParam)
			if err == nil {
				filter["bpm"] = bpm
			}
		}

		// if queryParam == "" {
		// 	log.Println("запрос пустой")
		// 	c.Header("Content-Type", "application/json")
		// 	c.JSON(http.StatusNotFound, gin.H{"Error": "ничего не получилось найти"})
		// 	c.Abort()
		// }

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		searchQueryDb, err := ProductCollection.Find(ctx, filter)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "something went wrong")
			return
		}
		err = searchQueryDb.All(ctx, &searchProducts)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusNotFound, "something went wrong")
			return
		}

		defer searchQueryDb.Close(ctx)

		if err := searchQueryDb.Err(); err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusNotFound, "invalid request")
			return
		}
		components.SoundList(searchProducts, userID).Render(c, c.Writer)
	}
}

func GenreToSlice(collection *mongo.Collection) ([]models.Genre, error) {
	var genres []models.Genre

	// Находим все документы в коллекции.
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var genre models.Genre
		if err := cur.Decode(&genre); err != nil {
			return nil, err
		}
		genres = append(genres, genre)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return genres, nil
}
func KeyToSlice(collection *mongo.Collection) ([]models.Key, error) {
	var keys []models.Key

	// Находим все документы в коллекции.
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var key models.Key
		if err := cur.Decode(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}
func InstrumentToSlice(collection *mongo.Collection) ([]models.Instrument, error) {
	var instruments []models.Instrument

	// Находим все документы в коллекции.
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var instrument models.Instrument
		if err := cur.Decode(&instrument); err != nil {
			return nil, err
		}
		instruments = append(instruments, instrument)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return instruments, nil
}
func VariationsToSlice(collection *mongo.Collection) ([]models.InstrumentVariation, error) {
	var variations []models.InstrumentVariation

	// Находим все документы в коллекции.
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var variation models.InstrumentVariation
		if err := cur.Decode(&variation); err != nil {
			return nil, err
		}
		variations = append(variations, variation)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return variations, nil
}
func ProductssToSlice(collection *mongo.Collection) ([]models.Product, error) {
	var products []models.Product

	// Находим все документы в коллекции.
	cur, err := collection.Find(context.Background(), bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(context.Background())

	for cur.Next(context.Background()) {
		var product models.Product
		if err := cur.Decode(&product); err != nil {
			return nil, err
		}
		products = append(products, product)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	return products, nil
}
func HomePage() gin.HandlerFunc {
	return func(c *gin.Context) {
		views.Home().Render(c, c.Writer)
	}
}
func ProfilePage() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("uid")
		if userID == "" {
			fmt.Println("нет такого")
		}

		products, err := ProductssToSlice(ProductCollection)
		if err != nil {
			log.Println(err)
			return
		}

		var user models.User
		err = UserCollection.FindOne(c, bson.M{"user_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		}
		fmt.Printf("User ID: %s\n", user.ID.Hex())
		fmt.Printf("First Name: %s\n", *user.First_Name)
		fmt.Printf("Last Name: %s\n", *user.Last_Name)
		fmt.Printf("Email: %s\n", *user.Email)
		fmt.Printf("Phone: %s\n", *user.Phone)
		views.ProfProd(products, &user).Render(c, c.Writer)
	}
}
func DeleteProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получение идентификатора продукта из параметра запроса
		productID := c.Query("id")
		if productID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID"})
			return
		}

		// Преобразование productID в ObjectID
		objID, err := primitive.ObjectIDFromHex(productID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product ID format"})
			return
		}

		filter := bson.M{"_id": objID}

		result, err := ProductCollection.DeleteOne(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
			return
		}
		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "product deleted successfully"})
	}
}
func GetSoundsPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDRaw, _ := c.Get("uid")
		if userIDRaw == "" {
			fmt.Println("нет такого")
			return
		}
		userID, ok := userIDRaw.(string)
		if !ok {
			fmt.Println("userID не является строкой")
			return
		}
		genres, err := GenreToSlice(genreCollection)
		if err != nil {
			log.Println("Failed to get genres:", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Internal Server Error"})
			return
		}
		keys, err := KeyToSlice(keyCollection)
		if err != nil {
			log.Println("Failed to get keys:", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Internal Server Error"})
			return
		}
		instruments, err := InstrumentToSlice(instrumentCollection)
		if err != nil {
			log.Println("Failed to get instruments:", err)
			c.HTML(http.StatusInternalServerError, "error.html", gin.H{"message": "Internal Server Error"})
			return
		}
		products, err := ProductssToSlice(ProductCollection)
		if err != nil {
			log.Println(err)
			return
		}
		views.Sounds(products, userID, genres, keys, instruments).Render(c, c.Writer)
	}
}
func GetCartPage() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDRaw, _ := c.Get("uid")
		if userIDRaw == "" {
			fmt.Println("нет такого")
		}
		userID, ok := userIDRaw.(string)
		if !ok {
			fmt.Println("userID не является строкой")
			return
		}
		var user models.User
		err := UserCollection.FindOne(c, bson.M{"user_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		}
		views.Cart(user.User_Cart, userID).Render(c, c.Writer)
	}
}
func UpdateTotalCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDRaw, _ := c.Get("uid")
		if userIDRaw == "" {
			fmt.Println("нет такого")
		}
		userID, ok := userIDRaw.(string)
		if !ok {
			fmt.Println("userID не является строкой")
			return
		}
		var user models.User
		err := UserCollection.FindOne(c, bson.M{"user_id": userID}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
		}

		components.CartTotal(user.User_Cart, userID).Render(c, c.Writer)
	}
}
