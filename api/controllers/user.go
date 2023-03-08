package controllers

import (
	"fmt"
	"net/http"
	"net/mail"
	"time"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/utils"
	"github.com/gin-gonic/gin"
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

	//generate code
	validationCode := utils.GetValidationCode()

	data := interfaces.User{Username: form.Username,
		Email:          form.Email,
		Password:       string(hashed),
		ValidationCode: validationCode,
		TimeExpiration: time.Now().Add(time.Minute * 60),
		IsVerified:     false}

	err = models.InsertUser(data)

	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	//send mail of verification
	err = utils.SendEmail(form.Email, "Here is your validation code", validationCode)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

func HandleCodeValidation(c *gin.Context) {
	var form interfaces.ValidationCode
	if err := c.BindJSON(&form); err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}
	//Check if there is no code
	if form.ValidationCode == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Verification code cannot be empty"})
		return
	}
	//Check if there is no email
	if form.Email == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Email cannot be empty"})
		return
	}
	// code validation
	exists := models.CheckCodeExistence(form.Email, form.ValidationCode)
	if exists {
		c.IndentedJSON(http.StatusOK, gin.H{"message": "Email has been verified"})
		return
	} else {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Code was incorrect or time has expired"})
		return
	}
}

func HandleNewCodeValidation(c *gin.Context) {
	var form interfaces.Email
	if err := c.BindJSON(&form); err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}
	//Check if there is no email
	if form.Email == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Verification code cannot be empty"})
		return
	}
	_, err := models.GetUserByEmail(form.Email)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "This Email has not been registered"})
		return
	}
	//generate code
	validationCode := utils.GetValidationCode()
	//update in database
	err = models.UpdateCode(form.Email, validationCode)
	if err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	//send mail of verification
	err = utils.SendEmail(form.Email, "Here is validation code requested", validationCode)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{"message": "New Code created and sended"})
}
