package controllers

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/mail"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func HandleSignUp(c *gin.Context) {
	var err error
	var form interfaces.SignUpForm

	if err := c.BindJSON(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	//Check if exists empty fields
	if form.Username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Username cannot be empty"})
		return
	}

	_, err = mail.ParseAddress(form.Email)

	if form.Email == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Email cannot be empty"})
		return
	} else if err != nil { //Check email format
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Email is invalid"})
		return
	}

	if form.Password == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Password cannot be empty"})
		return
	}

	//Check password format
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

	_, err = models.GetUserByEmail(form.Email)

	//Check if exists email
	if !(err != nil) {
		if !(err == mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(http.StatusConflict,
				gin.H{"message": "Email already exists"})
			return
		}
	}

	_, err = models.GetUserByUsername(form.Username)

	//Check if exists username
	if !(err != nil) {
		if !(err == mongo.ErrNoDocuments) {
			c.AbortWithStatusJSON(http.StatusConflict,
				gin.H{"message": "Username already exists"})
			return
		}
	}

	//encrypt password
	hashed, err := bcrypt.GenerateFromPassword([]byte(form.Password), 8)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	data := interfaces.User{Username: form.Username, Email: form.Email, Password: string(hashed), IsVerified: false}

	//Insert user in database
	var userId primitive.ObjectID
	userId, err = models.InsertUser(data)

	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	//generate code
	numbers := [...]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9"}
	var validationCode string = ""
	for i := 0; i < 6; i++ {
		randnum := rand.Intn(9)
		validationCode += numbers[randnum]
	}

	// updates validation code //todo ask
	err = models.InsertValidationCode(userId, validationCode)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	//send mail of verification, //todo validation? trycatch, revisar node
	//utils.SendEmail(verifcode)

	c.IndentedJSON(http.StatusOK, gin.H{"message": "User created successfully"})
}
