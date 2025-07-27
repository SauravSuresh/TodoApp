package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	db "github.com/SauravSuresh/persistence"
	"github.com/SauravSuresh/persistence/models"
	"github.com/SauravSuresh/todoapp/middlewares"
	"github.com/SauravSuresh/todoapp/services"
	"github.com/SauravSuresh/todoapp/utils"
	"github.com/go-chi/chi/v5"
	"github.com/thedevsaddam/renderer"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var rnd *renderer.Render

func init() {
	rnd = renderer.New(
		renderer.Options{
			ParseGlobPattern: "html/*.html",
		},
	)
}

type TodoHandlers struct {
	TodoSvc services.TodoService
}

// page handlers
func HomeHandler(rw http.ResponseWriter, r *http.Request) {
	// err := rnd.JSON(rw, http.StatusOK, "./readme.md")
	// utils.CheckErr(err, "failed to send response from home handler")
	err := rnd.HTML(rw, http.StatusOK, "registerPage", nil)
	utils.CheckErr(err, "failed to send response from home handler")
}

func IndexHandler(rw http.ResponseWriter, r *http.Request) {
	// err := rnd.JSON(rw, http.StatusOK, "./readme.md")
	// utils.CheckErr(err, "failed to send response from home handler")
	raw, ok := middlewares.GetUserID(r)
	if !ok {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{
			"message": "not logged in",
		})
		return
	}

	user, ok := raw.(*models.UserModel)
	if !ok {
		// fallback: we only have the ID string
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{
			"message": "user not loaded",
		})
		return
	}

	data := renderer.M{
		"Username": user.Username, // exported field
	}
	rw.Header().Set("Cache-Control", "no-store")
	if err := rnd.HTML(rw, http.StatusOK, "indexPage", data); err != nil {
		utils.CheckErr(err, "failed to send response from home handler")
	}
}

// todo handlers
func (t *TodoHandlers) CreateTodoHandler(rw http.ResponseWriter, r *http.Request) {
	var req models.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{
			"message": "Failed to decode JSON",
			"error":   err.Error(),
		})
		return
	}

	uidHex, err := utils.UserIDFromContext(r)
	if err != nil {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{"message": err.Error()})
		return
	}

	id, err := t.TodoSvc.Create(r.Context(), req, uidHex)

	if err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to add todo to db",
			"error":   err.Error(),
		})
		return
	}
	rnd.JSON(rw, http.StatusOK, renderer.M{
		"message": "Todo created successfully",
		"ID":      id,
	})

}

func (t *TodoHandlers) GetTodoHandler(rw http.ResponseWriter, r *http.Request) {
	var todoList []models.Todo
	todoList, err := t.TodoSvc.Get(r.Context(), map[string]interface{}{}, r)
	if err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not get todos",
			"error":   err.Error(),
		})
	}
	rnd.JSON(rw, http.StatusOK, models.GetObjectResponse{
		Message: "All Todos retrieved",
		Data:    todoList,
	})
}

func (t *TodoHandlers) GetCreatedTodoHandler(rw http.ResponseWriter, r *http.Request) {
	uid, err := utils.UserIDFromContext(r)
	if err != nil {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{"message": err.Error()})
		return
	}

	var todoList []models.Todo
	todoList, err = t.TodoSvc.Get(r.Context(), map[string]interface{}{"createdby": uid}, r)
	if err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not get todos",
			"error":   err.Error(),
		})
	}
	rnd.JSON(rw, http.StatusOK, models.GetObjectResponse{
		Message: "All Todos retrieved",
		Data:    todoList,
	})
}

func (t *TodoHandlers) GetAssignedTodoHandler(rw http.ResponseWriter, r *http.Request) {
	uid, err := utils.UserIDFromContext(r)
	if err != nil {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{"message": err.Error()})
		return
	}

	var todoList []models.Todo
	todoList, err = t.TodoSvc.Get(
		r.Context(),
		map[string]interface{}{
			"assignedto": uid,
			"accepted":   true,
		},
		r,
	)
	if err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not get todos",
			"error":   err.Error(),
		})
	}
	rnd.JSON(rw, http.StatusOK, models.GetObjectResponse{
		Message: "All Todos retrieved",
		Data:    todoList,
	})
}

func (t *TodoHandlers) GetInboxTodoHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Println("sdadas")
	uid, err := utils.UserIDFromContext(r)
	if err != nil {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{"message": err.Error()})
		return
	}

	var todoList []models.Todo
	todoList, err = t.TodoSvc.Get(
		r.Context(),
		map[string]interface{}{
			"assignedto": uid,
			"accepted":   false,
		},
		r,
	)
	if err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not get todos",
			"error":   err.Error(),
		})
	}
	log.Printf("GetInboxTodoHandler — retrieved %d todos", len(todoList))
	if payload, mErr := json.Marshal(todoList); mErr != nil {
		log.Printf("GetInboxTodoHandler — failed to marshal todos: %v", mErr)
	} else {
		log.Printf("GetInboxTodoHandler — todos: %s", string(payload))
	}
	rnd.JSON(rw, http.StatusOK, models.GetObjectResponse{
		Message: "All Todos retrieved",
		Data:    todoList,
	})
}

func (t *TodoHandlers) DeleteTodoHandler(rw http.ResponseWriter, r *http.Request) {

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	res, err := primitive.ObjectIDFromHex(id)
	fmt.Println(res.String())
	if err != nil {
		log.Printf("invalid id: %v\n", err.Error())
		rnd.JSON(rw, http.StatusBadRequest, err.Error())
		return
	}
	data, err := t.TodoSvc.Delete(r.Context(), res)
	if err != nil {
		log.Printf("could not delete item from database: %v\n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "an error occurred while deleting todo item",
			"error":   err.Error(),
		})
	} else {
		rnd.JSON(rw, http.StatusOK, renderer.M{
			"message": "item deleted successfully",
			"data":    data,
		})
	}
}

func (t *TodoHandlers) UpdateTodoHandler(rw http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("invalid hex id %q: %v", id, err)
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
			"error":   err.Error(),
		})
		return
	}

	var p models.UpdatePayload
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		log.Printf("decode error: %v", err)
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{
			"message": "Failed to decode JSON",
			"error":   err.Error(),
		})
		return
	}
	log.Printf("UpdateTodoHandler — payload: %+v", p)

	result, err := t.TodoSvc.Update(r.Context(), oid, p)

	if err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "failed to update todo",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("matched=%d modified=%d", result.MatchedCount, result.ModifiedCount)

	rnd.JSON(rw, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
		"data":    result.ModifiedCount,
	})
}

// user handlers

func GetAvaialableUsers(rw http.ResponseWriter, r *http.Request) {
	var UserListFromDB []models.UserModel
	filter := bson.M{}
	cursor, err := db.Db.Collection(db.GetUserCollectionName()).Find(context.Background(), filter)
	if err != nil {
		log.Printf("Failed to get users from db")
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not fetch users",
			"error":   err.Error(),
		})
		return
	}
	userList := []models.User{}
	if err := cursor.All(context.Background(), &UserListFromDB); err != nil {
		log.Printf("failed to extract from cursor %v \n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not extract users from cursor",
			"error":   err.Error(),
		})
	}
	for _, td := range UserListFromDB {
		userList = append(userList, td.ToUser())
	}
	rnd.JSON(rw, http.StatusOK, models.GetObjectResponse{
		Message: "All Users retrieved",
		Data:    userList,
	})
}

func RegisterUserHandler(rw http.ResponseWriter, r *http.Request) {
	var newUserFromRequest models.User
	if err := json.NewDecoder(r.Body).Decode(&newUserFromRequest); err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to decode JSON",
			"error":   err,
		})
		return
	}
	newUsertoDb := newUserFromRequest.ToUserModel()
	id, err := utils.MaybeAddUser(newUsertoDb, r)
	if err != nil {
		if err.Error() == "user already exists" {
			rnd.JSON(rw, http.StatusConflict, renderer.M{ // 409
				"message": err.Error(),
			})
		} else {
			rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
				"message": "DB error",
				"error":   err.Error(),
			})
		}
		return
	}

	//TODO move to env variable
	tokenstring, tokenerr := utils.GenerateToken(newUsertoDb.ID)
	if tokenerr != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "failed to generate token",
			"err":     err.Error(),
		})
		return
	}
	// Set JWT as secure, HTTP‑only cookie

	http.SetCookie(rw, utils.AddAuthCookie(tokenstring))

	rnd.JSON(rw, http.StatusOK, renderer.M{
		"message": "user created successfully",
		"ID":      id,
	})

}

func LoginPageHandler(rw http.ResponseWriter, r *http.Request) {
	// err := rnd.JSON(rw, http.StatusOK, "./readme.md")
	// utils.CheckErr(err, "failed to send response from home handler")
	err := rnd.HTML(rw, http.StatusOK, "loginPage", nil)
	utils.CheckErr(err, "failed to send response from home handler")
}

func Logout(rw http.ResponseWriter, r *http.Request) {

	expired := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0), // already in the past
		MaxAge:   -1,              // force removal in all browsers
		HttpOnly: true,
		Secure:   false, // flip to true when you serve over HTTPS
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(rw, expired)

	http.Redirect(rw, r, "/auth/login", http.StatusSeeOther)
}

func LoginAttemptHandler(rw http.ResponseWriter, r *http.Request) {
	var userfromRequest models.User
	fmt.Println(r.Body)
	if err := json.NewDecoder(r.Body).Decode(&userfromRequest); err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to decode JSON",
			"error":   err,
		})
		return
	}
	if userfromRequest.Email == "" || userfromRequest.Password == "" {
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{
			"message": "email and password cannot be empty",
		})
		return
	}

	var userFromDB models.UserModel
	err := db.Db.Collection(db.GetUserCollectionName()).FindOne(
		r.Context(),
		bson.M{"email": userfromRequest.Email},
	).Decode(&userFromDB)

	if err != nil {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{
			"message": "User Not found",
		})
		return
	}

	err = utils.ComparePassword(userFromDB.Password, userfromRequest.Password)
	if err != nil {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{
			"message": "Incorrect Password",
		})
		return
	}

	tokenstring, tokenerr := utils.GenerateToken(userFromDB.ID)
	if tokenerr != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "failed to generate token",
			"err":     err.Error(),
		})
		return
	}
	// Set JWT as secure, HTTP‑only cookie

	http.SetCookie(rw, utils.AddAuthCookie(tokenstring))
	http.Redirect(rw, r, "/todo/index", http.StatusSeeOther)
}
