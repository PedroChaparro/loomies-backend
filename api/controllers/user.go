package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const uri = "mongodb://root:development@localhost:27017/"

func Pruebas() {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))

	if err != nil {
		fmt.Println("Mongo.connect() error: ", err)
		os.Exit(1)
	}
	// Obtain the DB, by name. db will have the type
	// *mongo.Database
	db := client.Database("loomies")

	// use a filter to only select capped collections
	command := bson.D{{"create", "users"}}
	var result bson.M
	if err := db.RunCommand(context.TODO(), command).Decode(&result); err != nil {
		log.Fatal(err)
	}
}

func HandleSignin(c *gin.Context) {
	//Pruebas()
	var err error
	var form interfaces.SigninForm

	if err := c.BindJSON(&form); err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	err = models.CheckExistEmail(form.Email)

	if !(err != nil) {
		if !(err == mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"message": "Email already exists"})
			return
		}
	}

	err = models.CheckExistUser(form.User)

	if !(err != nil) {
		if !(err == mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"message": "User already exists"})
			return
		}
	}

	if form.User == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "User cannot be empty"})
		return
	}

	_, err = mail.ParseAddress(form.Email)

	if form.Email == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Email cannot be empty"})
		return
	} else if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Email is invalid"})
		return
	}

	if form.Password == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Password cannot be empty"})
		return
	}

	if len(form.Password) >= 8 {
		message := models.ValidPassword(form.Password)
		if message != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": message.Error()})
			return
		}
	} else {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Password is too short"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(form.Password), 8)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	data := map[string]interface{}{
		"user":     form.User,
		"email":    form.Email,
		"password": string(hashed),
		"items":    [][]string{},
		"loomies":  []string{},
	}

	err = models.InsertUser(data)

	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "User created successfully"})
}
