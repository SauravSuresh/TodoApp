package common

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	dbname          = "todo-database"
	todo_collection = "todo"
	user_collection = "users"
)

func GetTodoCollectionName() string {
	return todo_collection
}
func GetUserCollectionName() string {
	return user_collection
}

func GetDbName() string {
	return dbname
}

var Client *mongo.Client
var Db *mongo.Database

type (
	TodoModel struct {
		ID        primitive.ObjectID `bson:"_id,omitempty"`
		Title     string             `bson:"title"`
		CreatedAt primitive.DateTime `bson:"string"`
		DueDate   primitive.DateTime `bson:"duedate"`
		Completed bool               `bson:"completed"`
	}

	Todo struct {
		ID        string             `json:"id,omitempty"`
		Title     string             `json:"title"`
		CreatedAt primitive.DateTime `json:"string"`
		DueDate   primitive.DateTime `json:"duedate"`
		Completed bool               `json:"completed"`
	}

	GetTodoResponse struct {
		Message string `json:"message"`
		Data    []Todo `json:"data"`
	}

	UserModel struct {
		ID       primitive.ObjectID `bson:"id"`
		Username string             `bson:"username"`
		Email    string             `bson:email`
		Password string             `bson:"password";`
	}

	User struct {
		ID       string `json:"id,omitempty"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
)

func (tm TodoModel) ToTodo() Todo {
	return Todo{
		ID:        tm.ID.Hex(),
		Title:     tm.Title,
		CreatedAt: tm.CreatedAt,
		DueDate:   tm.DueDate,
		Completed: tm.Completed,
	}
}

func (tm Todo) ToTodoModel() TodoModel {
	objectID, err := primitive.ObjectIDFromHex(tm.ID)
	if err != nil {
		objectID = primitive.NewObjectID()
	}
	return TodoModel{
		ID:        objectID,
		Title:     tm.Title,
		CreatedAt: tm.CreatedAt,
		DueDate:   tm.DueDate,
		Completed: tm.Completed,
	}
}

func (u User) ToUserModel() UserModel {
	objectID, err := primitive.ObjectIDFromHex(u.ID)
	if err != nil {
		objectID = primitive.NewObjectID()
	}
	return UserModel{
		ID:       objectID,
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
	}
}
