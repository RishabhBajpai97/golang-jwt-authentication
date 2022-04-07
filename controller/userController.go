package controller

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/RishabhBajpai97/golang-jwt-authentication/database"
	"github.com/RishabhBajpai97/golang-jwt-authentication/helpers"
	"github.com/RishabhBajpai97/golang-jwt-authentication/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string{
  bytes,err := bcrypt.GenerateFromPassword([]byte(password),14)
	if err!=nil{
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, foundUserPassword string,)(bool,string) {
		err:=bcrypt.CompareHashAndPassword([]byte(foundUserPassword),[]byte(userPassword))
		var check=true
		var msg=""
		if err!=nil{
			msg = fmt.Sprintf("Incorrect email or password")
			check=false
		}
		return check,msg
}

func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationError := validate.Struct(user)
		if validationError != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationError.Error()})
			return
		}
		emailCount, err := userCollection.CountDocuments(context.Background(), bson.M{"email": user.Email})
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while checking the existing accounts"})
		}
		password:=HashPassword(*user.Password)
		user.Password = &password
		phoneCount, err := userCollection.CountDocuments(context.Background(), bson.M{"phone": user.Phone})

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while checking for phone number"})
		}
		if emailCount > 0 || phoneCount > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "This email or phone number already exists"})
		}
		user.Created_at,_= time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		user.Updated_at,_= time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserID = user.ID.Hex()
		token,refreshToken,_:= helpers.GenerateAllTokens(*user.Email,*user.First_name,*user.Last_name,*user.User_type,*&user.UserID)
		user.Token = &token
		user.Refresh_token = &refreshToken

		resultInsertionNumber,insertionErr:= userCollection.InsertOne(context.Background(), user)
		if insertionErr!=nil{
			msg:= fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error":msg})
			return
		}
		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}
func Login() gin.HandlerFunc {
		return func(c *gin.Context) {
			var user models.User
			var foundUser models.User
			if err:=c.BindJSON(&user); err!=nil{
				c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
				return
			}
			if err:=userCollection.FindOne(context.Background(), bson.M{"email":user.Email}).Decode(&foundUser);err!=nil{
					c.JSON(http.StatusInternalServerError, gin.H{"erorr":"email or password is incorrect"})
					return 
			}

			passwordIsValid,msg:=VerifyPassword(*user.Password, *foundUser.Password)
			if passwordIsValid!=true{
				c.JSON(http.StatusBadRequest,gin.H{"erorr":msg})
			}
			if foundUser.Email ==nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error":"user not found"})
			}
			token,refreshToken,_:= helpers.GenerateAllTokens(*foundUser.Email,*foundUser.First_name,*foundUser.Last_name,*foundUser.User_type,*&foundUser.UserID)
			helpers.UpdateAllTokens(token,refreshToken,foundUser.UserID)
			err:=userCollection.FindOne(context.Background(),bson.M{"user_id":foundUser.UserID}).Decode(&foundUser)
			if err!=nil{
				c.JSON(http.StatusInternalServerError, gin.H{"error":err.Error()})
				return
			}
			c.JSON(http.StatusOK, foundUser)
		}
}

func GetUsers() gin.HandlerFunc{
return func(c *gin.Context) {
		if err := helpers.CheckUserType(c, "ADMIN"); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}
		
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage <1{
			recordPerPage = 10
		}
		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 !=nil || page<1{
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))

		matchStage := bson.D{{"$match", bson.D{{}}}}
		groupStage := bson.D{{"$group", bson.D{
			{"_id", bson.D{{"_id", "null"}}}, 
			{"total_count", bson.D{{"$sum", 1}}}, 
			{"data", bson.D{{"$push", "$$ROOT"}}}}}}
		projectStage := bson.D{
			{"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"user_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},}}}
result,err := userCollection.Aggregate(context.Background(), mongo.Pipeline{
	matchStage, groupStage, projectStage})
if err!=nil{
	c.JSON(http.StatusInternalServerError, gin.H{"error":"error occured while listing user items"})
}
var allusers []bson.M
if err = result.All(context.Background(), &allusers); err!=nil{
	log.Fatal(err)
}
c.JSON(http.StatusOK, allusers[0])
	}
}

func GetUserById() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Param("user_id")
		if err := helpers.MatchUserTypeToUid(c, userId); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var user models.User
		err := userCollection.FindOne(context.Background(), bson.M{"user_id": userId}).Decode(&user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, user)
	}
}
