package controllers

import (
	"fmt"
	"net/http"
	"net/mail"

	"github.com/PedroChaparro/loomies-backend/interfaces"
	"github.com/PedroChaparro/loomies-backend/models"
	"github.com/PedroChaparro/loomies-backend/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

func HandleLogIn(c *gin.Context) {
	var err error
	var form interfaces.LogInForm
	var user interfaces.User

	if err := c.BindJSON(&form); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad request"})
		return
	}

	_, err = mail.ParseAddress(form.Email)

	//Check if exists empty fields
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

	user, err = models.GetUserByEmail(form.Email)

	//Check if exists email
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"message": "Wrong Email/Password"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				gin.H{"message": "Internal server error"})
			return
		}
	}

	//Check if the password is correct
	if err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized,
			gin.H{"message": "Wrong Email/Password"})
		return
	}

	//Check if user is verified
	if !user.IsVerified {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"message": "User has not been verified"})
		return
	}

	accessToken, err := utils.CreateAccessToken(user.Id.Hex())
	refreshToken, err := utils.CreateRefreshToken(user.Id.Hex())

	c.IndentedJSON(http.StatusOK, gin.H{
		"message":      "Successfully logged in",
		"user":         gin.H{"username": user.Username, "email": user.Email},
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	})
}

// HandleWhoami returns the user information (to recover the user session in the frontend)
func HandleWhoami(c *gin.Context) {
	userid, _ := c.Get("userid")
	user, err := models.GetUserById(userid.(string))

	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "User was not found"})
			return
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
			return
		}
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"message": "Successfully retrieved user",
		"user":    gin.H{"username": user.Username, "email": user.Email},
	})
}

// HandleRefresh returns a new access token
func HandleRefresh(c *gin.Context) {
	userid, _ := c.Get("userid")

	accessToken, err := utils.CreateAccessToken(userid.(string))
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	c.IndentedJSON(http.StatusOK, gin.H{
		"message":     "Successfully refreshed access token",
		"accessToken": accessToken,
	})
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
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "Code was incorrect"})
		return
	}
}
