package handlers

import (
	"fmt"
	"github.com/danielstutzman/todomvc-backend-go/models"
	"log"
	"strconv"
)

type Body struct {
	// ResetModel is for testing purposes
	ResetModel    bool                  `json:"resetModel"`
	DeviceUid     string                `json:"deviceUid"`
	ActionsToSync []models.ActionToSync `json:"actionsToSync"`
}

type Response struct {
	DeviceId               int            `json:"deviceId"`
	ActionToSyncIdToOutput map[string]int `json:"actionToSyncIdToOutput"`
	Todos                  []models.Todo  `json:"todos"`
}

func mapIntIntToMapStringInt(input map[int]int) map[string]int {
	output := map[string]int{}
	for k, v := range input {
		output[strconv.Itoa(k)] = v
	}
	return output
}

func HandleBody(body Body, model models.Model) (*Response, error) {
	log.Printf("-- Got body %v", body)

	if body.ResetModel {
		model.Reset()
	}

	if body.DeviceUid == "" {
		return nil, fmt.Errorf("Blank DeviceUid")
	}
	device := model.FindOrCreateDeviceByUid(body.DeviceUid)
	log.Println("   Got device", device)

	tempIdToId := map[int]int{}
	for _, actionToSync := range body.ActionsToSync {
		_, alreadyExecuted := device.ActionToSyncIdToOutput[actionToSync.Id]
		if !alreadyExecuted {
			output, err := handleActionToSync(actionToSync, model, tempIdToId)
			if err != nil {
				return nil, fmt.Errorf("Error from handleActionToSync: %s", err)
			}

			// immutable edit so we don't corrupt MemoryModel
			device.ActionToSyncIdToOutput = immutableSetForMapIntInt(
				device.ActionToSyncIdToOutput, actionToSync.Id, output)
		}

		if actionToSync.Type == "TODOS/ADD_TODO" {
			tempIdToId[actionToSync.TodoIdMaybeTemp] =
				device.ActionToSyncIdToOutput[actionToSync.Id]
		}
	}
	model.UpdateDeviceActionToSyncIdToOutputJson(device)

	response := Response{
		DeviceId:               device.Id,
		ActionToSyncIdToOutput: mapIntIntToMapStringInt(device.ActionToSyncIdToOutput),
		Todos: model.ListTodos(),
	}
	return &response, nil
}

// returns output -- the new TodoID if TODOS/ADD_TODOS, the number of rows updated
// for other types
func handleActionToSync(actionToSync models.ActionToSync,
	model models.Model, tempIdToId map[int]int) (int, error) {

	var todoId int
	if actionToSync.Type != "TODOS/ADD_TODO" {
		if actionToSync.TodoIdMaybeTemp < 0 {
			var ok bool
			todoId, ok = tempIdToId[actionToSync.TodoIdMaybeTemp]
			if !ok {
				return 0,
					fmt.Errorf("Don't know todoId for temp id in action %v", actionToSync)
			}
		} else if actionToSync.TodoIdMaybeTemp > 0 {
			todoId = actionToSync.TodoIdMaybeTemp
		} else {
			return 0, fmt.Errorf("Invalid TodoIdMaybeTemp in action %v", actionToSync)
		}
	}

	switch actionToSync.Type {

	case "TODOS/ADD_TODO":
		log.Printf("  Calling CreateTodo(%v)", actionToSync)
		todo := model.CreateTodo(actionToSync)
		return todo.Id, nil

	case "TODO/UPDATE_TODO":
		log.Printf("  Calling UpdateTodo(%v)", actionToSync)
		output := model.UpdateTodo(actionToSync, todoId)
		return output, nil

	case "TODOS/DELETE_TODO":
		log.Printf("  Calling DeleteTodo(%v)", actionToSync)
		output := model.DeleteTodo(todoId)
		return output, nil

	default:
		return 0, fmt.Errorf("Unknown type in actionToSync: %v", actionToSync)
	}
}

// Set input[keyToSet] = valueToSet in a copy of input (doesn't modify input)
func immutableSetForMapIntInt(input map[int]int, keyToSet int,
	valueToSet int) map[int]int {
	output := map[int]int{}
	for k, v := range input {
		output[k] = v
	}
	output[keyToSet] = valueToSet
	return output
}
