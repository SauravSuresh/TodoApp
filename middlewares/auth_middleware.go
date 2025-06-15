package middlewares

import (
	"context"
	"net/http"

	"github.com/SauravSuresh/todoapp/common"
	"github.com/SauravSuresh/todoapp/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AuthenticationMiddelware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		var err error
		if c, err := r.Cookie("auth_token"); err == nil {
			tokenString = c.Value
		}
		if err != nil {
			http.Error(w, "no auth token found", http.StatusUnauthorized)
			return
		}
		claims, err := utils.VerifyToken(tokenString)
		if err != nil {
			http.Error(w, "Invalid authentication token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), common.UserIDKey, claims["user_id"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(r *http.Request) (interface{}, bool) {
	userID := r.Context().Value(common.UserIDKey)
	return userID, userID != nil
}

func UserLoaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := r.Context().Value(common.UserIDKey)
		oidHex, ok := userID.(string)
		if !ok {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		oid, err := primitive.ObjectIDFromHex(oidHex)
		if err != nil {
			http.Error(w, "invalid_user_id", http.StatusUnauthorized)
		}
		var u common.UserModel
		err = common.Db.Collection(common.GetUserCollectionName()).FindOne(r.Context(), bson.M{"id": oid}).Decode(&u)
		if err != nil {
			http.Error(w, "user not found", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), common.UserIDKey, &u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
