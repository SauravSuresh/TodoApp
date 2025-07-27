package services

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	database "github.com/SauravSuresh/persistence/interfaces"
	"github.com/SauravSuresh/persistence/models"
	"github.com/SauravSuresh/todoapp/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TodoService struct {
	repo database.TodoRepository
}

func NewTodoService(r database.TodoRepository) *TodoService {
	return &TodoService{repo: r}
}

func (s *TodoService) Create(ctx context.Context, req models.CreateTodoRequest, uid primitive.ObjectID) (primitive.ObjectID, error) {

	var dueDatePtr primitive.DateTime
	if req.DueDateMs > 0 {
		startOfToday := time.Now().Truncate(24 * time.Hour).UnixMilli()
		if req.DueDateMs < startOfToday {
			return primitive.NilObjectID, fmt.Errorf("Due date cannot be before today")
		}
		dt := primitive.DateTime(req.DueDateMs)
		dueDatePtr = dt
	}
	todo := &models.Todo{
		Title:      req.Title,
		DueDate:    dueDatePtr,
		AssignedTo: req.AssignedTo,
	}
	todomodel := todo.ToTodoModel()
	todomodel.CreatedAt = primitive.NewDateTimeFromTime(time.Now())
	log.Printf("Set CreatedAt to: %v", todomodel.CreatedAt.Time())
	todomodel.DueDate = primitive.DateTime(req.DueDateMs)
	todomodel.CreatedBy = uid

	if req.AssignedTo == "" {
		todomodel.AssignedTo = uid
	}

	return s.repo.Create(ctx, todomodel)
}

func (s *TodoService) Get(ctx context.Context, filter map[string]interface{}, r *http.Request) ([]models.Todo, error) {
	var bsonFilter interface{}

	// Build the filter from the provided map
	if len(filter) == 0 {
		bsonFilter = bson.D{}
	} else {
		var doc bson.D
		for k, v := range filter {
			doc = append(doc, bson.E{Key: k, Value: v})
		}
		bsonFilter = doc
	}

	todofromdb, err := s.repo.Get(ctx, bsonFilter)
	if err != nil {
		return nil, err
	}

	todoList := []models.Todo{}
	for _, td := range todofromdb {
		createdbyname, _ := utils.GetusernameFromID(td.CreatedBy, r)
		assignbyname, _ := utils.GetusernameFromID(td.AssignedTo, r)
		todoList = append(todoList, td.ToTodo(createdbyname, assignbyname))
	}

	return todoList, nil
}

func (s *TodoService) Delete(ctx context.Context, id primitive.ObjectID) (*mongo.DeleteResult, error) {
	filter := bson.M{"_id": id}
	data, err := s.repo.Delete(ctx, filter)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *TodoService) Update(ctx context.Context, id primitive.ObjectID, updateObj models.UpdatePayload) (*mongo.UpdateResult, error) {
	updateFields := bson.M{}
	if updateObj.Title != nil {
		title := strings.TrimSpace(*updateObj.Title)
		if title == "" {
			return nil, fmt.Errorf("Title cannot be empty")
		}
		updateFields["title"] = title
	}
	if updateObj.Completed != nil {
		updateFields["completed"] = updateObj.Completed
	}
	if updateObj.DueDateMs != nil {
		updateFields["duedate"] = primitive.DateTime(*updateObj.DueDateMs)
	}
	if len(updateFields) == 0 {
		return nil, fmt.Errorf("No fields to udpate")
	}

	updatefilter := bson.M{"_id": id}
	update := bson.M{"$set": updateFields}

	log.Printf("filter: %+v update: %+v", updatefilter, update)

	result, err := s.repo.Update(ctx, updatefilter, update)
	if err != nil {
		return nil, err
	}
	return result, nil
}
