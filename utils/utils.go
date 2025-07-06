package utils

import (
	"fmt"
	"log"
	"net/http"
	"time"

	db "github.com/SauravSuresh/persistence"
	"github.com/SauravSuresh/persistence/models"
	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var secretKey = []byte("secretpassword")

func GenerateToken(userID primitive.ObjectID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.Hex(),                     // store hex string
		"exp":     time.Now().Add(time.Hour).Unix(), // +1 h
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
func VerifyToken(tokenstring string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenstring, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("invalid signing method")
		}
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

func ComparePassword(A string, B string) error {
	if A == B {
		return nil
	}
	return fmt.Errorf("passwords dont match")
}

// TODO: hash passwords
func MaybeAddUser(newuser models.UserModel, r *http.Request) (interface{}, error) {
	start := time.Now()
	fmt.Println(newuser.Email)
	result := db.Db.Collection(db.GetUserCollectionName()).FindOne(r.Context(), bson.M{"email": newuser.Email})
	log.Printf("find-one took %v (err=%v)", time.Since(start), result.Err())
	if err := result.Err(); err == nil {
		return nil, fmt.Errorf("user already exists")
	}
	data, err := db.Db.Collection(db.GetUserCollectionName()).InsertOne(r.Context(), newuser)
	if err != nil {
		return err, nil
	}
	return data.InsertedID, nil
}

func AddAuthCookie(tokenstring string) *http.Cookie {
	return &http.Cookie{
		Name:     "auth_token",
		Value:    tokenstring,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour),
		HttpOnly: true,
		Secure:   false, // change to true when served over HTTPS
		SameSite: http.SameSiteLaxMode,
	}
}

type contextKey string

const userKey contextKey = "user"

func GetUser(r *http.Request) (*models.UserModel, bool) {
	usr, ok := r.Context().Value(userKey).(*models.UserModel)
	return usr, ok
}

// todo a fucntion that gets the username from ID
func GetusernameFromID(userID primitive.ObjectID, r *http.Request) (string, error) {
	var userfromDB models.UserModel
	result := db.Db.Collection((db.GetUserCollectionName())).FindOne(r.Context(), bson.M{"id": userID})
	if err := result.Err(); err != nil {
		return "", err
	}
	if err := result.Decode(&userfromDB); err != nil {
		return "", err
	}
	return userfromDB.Username, nil
}

func UserIDFromContext(r *http.Request) (primitive.ObjectID, error) {
	switch v := r.Context().Value(db.UserIDKey).(type) {
	case string: // raw hex string set by AuthenticationMiddleware
		return primitive.ObjectIDFromHex(v)
	case *models.UserModel: // already loaded by UserLoaderMiddleware
		return v.ID, nil
	default:
		return primitive.NilObjectID, fmt.Errorf("unauthenticated")
	}
}

func CheckErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s %v \n", msg, err)
	}
}
