package common

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type contextKey string

const UserIDKey contextKey = "user_id"

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
		ID         primitive.ObjectID `bson:"_id,omitempty"`
		Title      string             `bson:"title"`
		CreatedAt  primitive.DateTime `bson:"string"`
		DueDate    primitive.DateTime `bson:"duedate"`
		Completed  bool               `bson:"completed"`
		CreatedBy  primitive.ObjectID `bson:"createdby"`
		AssignedTo primitive.ObjectID `bson:"assignedto"`
	}

	CreateTodoRequest struct {
		Title      string `json:"title"`
		DueDateMs  int64  `json:"duedate"`    // epoch-milliseconds number
		AssignedTo string `json:"assignedto"` // hex user ID
	}

	SetStatusRequest struct {
		Id     string
		Update bool `json:"update"`
	}

	Todo struct {
		ID            string             `json:"id,omitempty"`
		Title         string             `json:"title"`
		CreatedAt     primitive.DateTime `json:"createdat"`
		DueDate       primitive.DateTime `json:"duedate"`
		Completed     bool               `json:"completed"`
		CreatedBy     string             `json:"createdby"`
		AssignedTo    string             `json:"assignedto"`
		CreatedByHex  string             `json:"createdbyhex"`
		AssignedToHex string             `json:"assignedtohex"`
	}

	GetObjectResponse struct {
		Message string      `json:"message"`
		Data    interface{} `json:"data"`
	}

	UserModel struct {
		ID       primitive.ObjectID `bson:"id"`
		Username string             `bson:"username"`
		Email    string             `bson:email`
		Password string             `bson:"password"`
	}

	User struct {
		ID       string `json:"id,omitempty"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
)

func (tm TodoModel) ToTodo(createdByUsername string, assignedToUsername string) Todo {
	return Todo{
		ID:         tm.ID.Hex(),
		Title:      tm.Title,
		CreatedAt:  tm.CreatedAt,
		DueDate:    tm.DueDate,
		Completed:  tm.Completed,
		CreatedBy:  createdByUsername,
		AssignedTo: assignedToUsername,
	}
}

func (tm Todo) ToTodoModel() TodoModel {
	objectID, err := primitive.ObjectIDFromHex(tm.ID)
	if err != nil {
		objectID = primitive.NewObjectID()
	}
	assignedObjectID, assingerr := primitive.ObjectIDFromHex(tm.AssignedTo)
	if assingerr != nil {
		objectID = primitive.ObjectID{}
	}
	return TodoModel{
		ID:         objectID,
		Title:      tm.Title,
		CreatedAt:  tm.CreatedAt,
		DueDate:    tm.DueDate,
		Completed:  tm.Completed,
		AssignedTo: assignedObjectID,
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

func (u UserModel) ToUser() User {
	return User{
		ID:       u.ID.Hex(),
		Username: u.Username,
		Email:    u.Email,
	}
}
