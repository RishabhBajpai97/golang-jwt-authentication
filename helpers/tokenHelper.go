package helpers

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/RishabhBajpai97/golang-jwt-authentication/database"
	jwt "github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Uid        string
	User_type  string
	jwt.StandardClaims
}

var userCollection = database.OpenCollection(database.Client, "user")
var SECRET_KEY = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstNAme string, lastName string, userType string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_name: firstNAme,
		Last_name:  lastName,
		Uid:        uid,
		User_type:  userType,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(24)).Unix(),
		},
	}
	refreshClaims := &SignedDetails{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Local().Add(time.Hour * time.Duration(168)).Unix(),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	refresToken,err:=jwt.NewWithClaims(jwt.SigningMethodHS256,refreshClaims).SignedString([]byte(SECRET_KEY))
	if err!=nil{
		log.Panic(err)
	}
	return token,refresToken,err
}


func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string)  {
	var updateObj primitive.D
	updateObj = append(updateObj, bson.E{"token",signedToken})
	updateObj = append(updateObj, bson.E{"refresh_token", signedRefreshToken})
	Updated_at,_ := time.Parse(time.RFC3339,time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{"updated_at", Updated_at})
	upsert:=true
	filter:=bson.M{"user_id":userId}
	opt:=options.UpdateOptions{
		Upsert: &upsert,
	}

	_,err:=userCollection.UpdateOne(context.Background(),filter,bson.D{{"$set",updateObj}},&opt)
	if err!=nil{
		log.Panic(err)
		return
	} 
	return
}
func ValidateToken(signedToken string) (claims *SignedDetails,  msg string) {
	token,err:= jwt.ParseWithClaims(signedToken,&SignedDetails{},func(t *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY),nil
	})
	if err!=nil{
		msg = err.Error()
		return
	}
	claims,ok := token.Claims.(*SignedDetails)
	if !ok{
		msg = fmt.Sprintf("the token is invalid")
		msg = err.Error()
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix(){
		msg = fmt.Sprintf("Token is expired")
		msg = err.Error()
	}
	return claims,msg
}