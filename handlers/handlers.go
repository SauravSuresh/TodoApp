package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/SauravSuresh/todoapp/common"
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
	err := rnd.HTML(rw, http.StatusOK, "indexPage", nil)
	utils.CheckErr(err, "failed to send response from home handler")
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
		todoList = append(todoList, td.ToTodo())
	}
	rnd.JSON(rw, http.StatusOK, common.GetTodoResponse{
		Message: "All Todos retrieved",
		Data:    todoList,
	})

}

func CreateTodoHandler(rw http.ResponseWriter, r *http.Request) {
	var newTodoFromRequest common.Todo
	if err := json.NewDecoder(r.Body).Decode(&newTodoFromRequest); err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to decode JSON",
			"error":   err,
		})
		return
	}
	newTodoForDB := newTodoFromRequest.ToTodoModel()

	data, err := common.Db.Collection(common.GetTodoCollectionName()).InsertOne(r.Context(), newTodoForDB)
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
		fmt.Println("OOMBI")
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to decode JSON",
			"error":   err,
		})
		return
	}
	newUsertoDb := newUserFromRequest.ToUserModel()
	data, err := common.Db.Collection(common.GetUserCollectionName()).InsertOne(r.Context(), newUsertoDb)
	if err != nil {
		rnd.JSON(rw, http.StatusInternalServerError, renderer.M{
			"message": "Failed to add user to db",
			"error":   err,
		})
		return
	}
	rnd.JSON(rw, http.StatusOK, renderer.M{
		"message": "user created successfully",
		"ID":      data.InsertedID,
	})

}

func LoginPageHandler(rw http.ResponseWriter, r *http.Request) {
	// err := rnd.JSON(rw, http.StatusOK, "./readme.md")
	// utils.CheckErr(err, "failed to send response from home handler")
	err := rnd.HTML(rw, http.StatusOK, "loginPage", nil)
	utils.CheckErr(err, "failed to send response from home handler")
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

	rnd.JSON(rw, http.StatusOK, renderer.M{
		"message": "Login successful",
		"ID":      userFromDB.ID,
	})
}
