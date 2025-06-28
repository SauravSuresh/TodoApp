package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/SauravSuresh/todoapp/common"
	"github.com/SauravSuresh/todoapp/middlewares"
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

	user, ok := raw.(*common.UserModel)
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

func GetTodoHandler(rw http.ResponseWriter, r *http.Request) {
	var todoListFromDB = []common.TodoModel{}
	filter := bson.D{}

	cursor, err := common.Db.Collection(common.GetTodoCollectionName()).Find(context.Background(), filter)
	if err != nil {
		log.Printf("failed to fetch todo records from db %v \n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not fetch the todo collection",
			"error":   err.Error(),
		})
		return
	}
	todoList := []common.Todo{}
	if err := cursor.All(context.Background(), &todoListFromDB); err != nil {
		log.Printf("failed to extract from cursor %v \n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not extract from cursor",
			"error":   err.Error(),
		})
	}

	for _, td := range todoListFromDB {
		createdbyname, _ := utils.GetusernameFromID(td.CreatedBy, r)
		assignbyname, _ := utils.GetusernameFromID(td.AssignedTo, r)
		fmt.Printf("assigned %s", assignbyname)
		todoList = append(todoList, td.ToTodo(createdbyname, assignbyname))
	}
	rnd.JSON(rw, http.StatusOK, common.GetObjectResponse{
		Message: "All Todos retrieved",
		Data:    todoList,
	})
}

func GetCreatedTodoHandler(rw http.ResponseWriter, r *http.Request) {
	var todoListFromDB = []common.TodoModel{}
	uid, err := utils.UserIDFromContext(r)
	if err != nil {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{"message": err.Error()})
		return
	}

	filter := bson.M{"createdby": uid}

	cursor, err := common.Db.Collection(common.GetTodoCollectionName()).Find(context.Background(), filter)
	if err != nil {
		log.Printf("failed to fetch todo records from db %v \n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not fetch the todo collection",
			"error":   err.Error(),
		})
		return
	}
	todoList := []common.Todo{}
	if err := cursor.All(context.Background(), &todoListFromDB); err != nil {
		log.Printf("failed to extract todos from cursor %v \n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not extract from cursor",
			"error":   err.Error(),
		})
	}

	for _, td := range todoListFromDB {
		createdbyname, _ := utils.GetusernameFromID(td.CreatedBy, r)
		assignbyname, _ := utils.GetusernameFromID(td.AssignedTo, r)
		fmt.Printf("assigned %s", assignbyname)
		todoList = append(todoList, td.ToTodo(createdbyname, assignbyname))
	}
	rnd.JSON(rw, http.StatusOK, common.GetObjectResponse{
		Message: "All Todos retrieved",
		Data:    todoList,
	})
}

func GetAssignedTodoHandler(rw http.ResponseWriter, r *http.Request) {
	var todoListFromDB = []common.TodoModel{}
	uid, err := utils.UserIDFromContext(r)
	if err != nil {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{"message": err.Error()})
		return
	}

	filter := bson.M{"assignedto": uid}

	cursor, err := common.Db.Collection(common.GetTodoCollectionName()).Find(context.Background(), filter)
	if err != nil {
		log.Printf("failed to fetch todo records from db %v \n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not fetch the todo collection",
			"error":   err.Error(),
		})
		return
	}
	todoList := []common.Todo{}
	if err := cursor.All(context.Background(), &todoListFromDB); err != nil {
		log.Printf("failed to extract todos from cursor %v \n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not extract from cursor",
			"error":   err.Error(),
		})
	}

	for _, td := range todoListFromDB {
		createdbyname, _ := utils.GetusernameFromID(td.CreatedBy, r)
		assignbyname, _ := utils.GetusernameFromID(td.AssignedTo, r)
		fmt.Printf("assigned %s", assignbyname)
		todoList = append(todoList, td.ToTodo(createdbyname, assignbyname))
	}
	rnd.JSON(rw, http.StatusOK, common.GetObjectResponse{
		Message: "All Todos retrieved",
		Data:    todoList,
	})
}

func GetAvaialableUsers(rw http.ResponseWriter, r *http.Request) {
	var UserListFromDB []common.UserModel
	filter := bson.M{}
	cursor, err := common.Db.Collection(common.GetUserCollectionName()).Find(context.Background(), filter)
	if err != nil {
		log.Printf("Failed to get users from db")
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Could not fetch users",
			"error":   err.Error(),
		})
		return
	}
	userList := []common.User{}
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
	rnd.JSON(rw, http.StatusOK, common.GetObjectResponse{
		Message: "All Users retrieved",
		Data:    userList,
	})
}

func CreateTodoHandler(rw http.ResponseWriter, r *http.Request) {

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("could not read body: %v\n", err)
		// you can bail out or just continue
	} else {
		fmt.Printf("RAW JSON >>> %s\n", string(bodyBytes))

		// Try unmarshalling into a generic map to inspect field types
		var dbg map[string]interface{}
		if err := json.Unmarshal(bodyBytes, &dbg); err == nil {
			for k, v := range dbg {
				fmt.Printf("field %-10s : %T (%v)\n", k, v, v)
			}
		} else {
			fmt.Printf("json.Unmarshal error: %v\n", err)
		}

		// Reset r.Body so the real decoder can read it again
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	}

	var req common.CreateTodoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{
			"message": "Failed to decode JSON",
			"error":   err.Error(),
		})
		return
	}

	dueDT := primitive.DateTime(req.DueDateMs)

	uidHex, err := utils.UserIDFromContext(r)
	if err != nil {
		rnd.JSON(rw, http.StatusUnauthorized, renderer.M{"message": err.Error()})
		return
	}

	todo := common.Todo{
		Title:      req.Title,
		DueDate:    dueDT,
		AssignedTo: req.AssignedTo,
	}

	model := todo.ToTodoModel()
	if err != nil {
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{"message": err.Error()})
		return
	}

	model.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	model.CreatedBy = uidHex
	data, err := common.Db.Collection(common.GetTodoCollectionName()).InsertOne(r.Context(), model)
	if err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to add todo to db",
			"error":   err,
		})
		return
	}
	rnd.JSON(rw, http.StatusOK, renderer.M{
		"message": "Todo created successfully",
		"ID":      data.InsertedID,
	})

}

func UpdateTodoHandler(rw http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(chi.URLParam(r, "id"))
	res, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("the id param is not a valid hex value: %v\n", err.Error())
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
			"error":   err.Error(),
		})
		return
	}
	var updateTodofromRequest common.Todo
	if err := json.NewDecoder(r.Body).Decode(&updateTodofromRequest); err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to decode JSON",
			"error":   err,
		})
		return
	}
	if updateTodofromRequest.Title == "" {
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{
			"message": "Title cannot be empty",
		})
		return
	}

	filter := bson.M{"id": res}
	update := bson.M{"$set": bson.M{
		"title":     updateTodofromRequest.Title,
		"completed": updateTodofromRequest.Completed,
	}}
	data, err := common.Db.Collection(common.GetTodoCollectionName()).UpdateOne(r.Context(), filter, update)
	if err != nil {
		log.Printf("failed to update db collection: %v\n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to update data in the database",
			"error":   err.Error(),
		})
		return
	}
	rnd.JSON(rw, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
		"data":    data.ModifiedCount,
	})

}

func DeleteTodoHandler(rw http.ResponseWriter, r *http.Request) {

	id := strings.TrimSpace(chi.URLParam(r, "id"))
	res, err := primitive.ObjectIDFromHex(id)
	fmt.Println(res.String())
	if err != nil {
		log.Printf("invalid id: %v\n", err.Error())
		rnd.JSON(rw, http.StatusBadRequest, err.Error())
		return
	}
	filter := bson.M{"_id": res}
	if data, err := common.Db.Collection(common.GetTodoCollectionName()).DeleteOne(r.Context(), filter); err != nil {
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

func RegisterUserHandler(rw http.ResponseWriter, r *http.Request) {
	var newUserFromRequest common.User
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
	return
}

func LoginAttemptHandler(rw http.ResponseWriter, r *http.Request) {
	var userfromRequest common.User
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

	var userFromDB common.UserModel
	err := common.Db.Collection(common.GetUserCollectionName()).FindOne(
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
	return
}

func SetStatusHandler(rw http.ResponseWriter, r *http.Request) {

	id := strings.TrimSpace(chi.URLParam(r, "id"))

	var updatereq common.SetStatusRequest
	res, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("the id param is not a valid hex value: %v\n", err.Error())
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{
			"message": "The id is invalid",
			"error":   err.Error(),
		})
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&updatereq); err != nil {
		rnd.JSON(rw, http.StatusBadRequest, renderer.M{
			"message": "Failed to decode JSON",
			"error":   err.Error(),
		})
		return
	}
	fmt.Printf("update from post is %v", updatereq.Update)

	filter := bson.M{"_id": res}
	update := bson.M{"$set": bson.M{
		"completed": updatereq.Update,
	}}
	// DEBUG: confirm we got the right ObjectID
	fmt.Printf("SetStatusHandler — resolved ObjectID: %s\n", res.Hex())

	data, err := common.Db.Collection(common.GetTodoCollectionName()).UpdateOne(r.Context(), filter, update)
	if err != nil {
		log.Printf("failed to update db collection: %v\n", err.Error())
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to update data in the database",
			"error":   err.Error(),
		})
		return
	}
	rnd.JSON(rw, http.StatusOK, renderer.M{
		"message": "Todo updated successfully",
		"data":    data.ModifiedCount,
	})

}
