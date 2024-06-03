package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID              primitive.ObjectID   `json:"_id" bson:"_id"`
	First_Name      *string              `json:"first_name" 					validate:"required, min=2",max=30 `
	Last_Name       *string              `json:"last_name" 						validate:"required, min=2",max=30`
	Password        *string              `json:"password" 						validate:"required, min=6" `
	Email           *string              `json:"email" 							validate:"email, required" `
	Phone           *string              `json:"phone" 							validate:"required"`
	Token           *string              `json:"token"`
	Refresh_Token   *string              `json:"refresh_token"`
	Created_At      time.Time            `json:"created_at"`
	Updated_At      time.Time            `json:"updated_at"`
	User_ID         string               `json:"user_id"`
	User_Cart       []Product_User       `json:"user_cart" bson:"user_cart"`
	Payment_Details []PaymentDetails     `json:"payment_details" bson:"payment_details"`
	Order_Status    []Order              `json:"orders" bson:"orders"`
	Uploaded_Files  []primitive.ObjectID `json:"uploaded_files" bson:"uploaded_files"`
}

type Product struct {
	Product_ID   primitive.ObjectID `bson:"_id"`
	Product_Name *string            `bson:"product_name"`
	Price        *uint64            `bson:"price"`
	Image        *string            `bson:"image"`
	Author_ID    primitive.ObjectID `bson:"author_id"`
	File_Link    *string            `bson:"file_link"`
	Genre        *[]string          `bson:"genre"`
	Bpm          *uint64            `bson:"bpm"`
	Key          *string            `bson:"key"`
	Instrument   *[]string          `bson:"instrument"`
	Variation    *[]string          `bson:"variation"`
}
type Genre struct {
	ID   primitive.ObjectID `bson:"_id"`
	Name string             `bson:"name"`
}
type Key struct {
	ID   primitive.ObjectID `bson:"_id"`
	Name string             `bson:"name"`
}
type Instrument struct {
	ID   primitive.ObjectID `bson:"_id"`
	Name string             `bson:"name"`
}
type InstrumentVariation struct {
	ID         primitive.ObjectID `bson:"_id"`
	Name       string             `bson:"name"`
	Instrument primitive.ObjectID `bson:"instrument"`
}
type Product_User struct {
	Product_ID   primitive.ObjectID `bson:"_id"`
	Product_Name *string            `bson:"product_name"`
	Price        int                `bson:"price"`
	Image        *string            `bson:"image"`
	Author_ID    primitive.ObjectID `bson:"author_id"`
	File_Link    *string            `bson:"file_link"`
	Genre        *[]string          `bson:"genre"`
	Bpm          *uint64            `bson:"bpm"`
	Key          *string            `bson:"key"`
	Instrument   *[]string          `bson:"instrument"`
	Variation    *[]string          `bson:"variation"`
}
type PaymentDetails struct {
	Address_ID primitive.ObjectID `bson:"_id"`
	CardNumber *string            `json:"card_number" bson:"card_number"`
}

type Order struct {
	Order_ID       primitive.ObjectID `bson:"_id"`
	Order_Cart     []Product_User     `json:"order_list" bson:"order_list"`
	Order_At       time.Time          `json:"order_at" bson:"order_at"`
	Price          int                `json:"total_price" bson:"total_price"`
	Discount       *int               `json:"discount" bson:"discount"`
	Payment_Method Payment            `json:"payment_method" bson:"payment_method"`
}

type Payment struct {
	Digital bool
	COD     bool
}

type ProductType struct {
	ID          primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	Properties  []string           `bson:"properties"`
}
