package common

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	dbname     = "todo-database"
	collection = "todo"
)

func GetCollectionName() string {
	return collection
}

func GetDbName() string {
	return dbname
}

var Client *mongo.Client
var Db *mongo.Database

type (
	TodoModel struct {
		ID        primitive.ObjectID `bson:"id,omitempty"`
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

	User struct {
		ID       uint   `json:"id"`
		Username string `json:"username"`
		Password string `json:"password";`
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
