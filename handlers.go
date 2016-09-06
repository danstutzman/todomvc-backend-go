package main

import (
	"./model"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Response struct {
	DeviceId               int            `json:"deviceId"`
	ActionToSyncIdToOutput map[string]int `json:"actionToSyncIdToOutput"`
	Todos                  []model.Todo   `json:"todos"`
}

func mapIntIntToMapStringInt(input map[int]int) map[string]int {
	output := map[string]int{}
	for k, v := range input {
		output[strconv.Itoa(k)] = v
	}
	return output
}

func handleRequest(writer http.ResponseWriter, request *http.Request,
	model model.Model) {
	// Set Access-Control-Allow-Origin for all requests
	writer.Header().Set("Access-Control-Allow-Origin", "*")

	switch request.Method {
	case "GET":
		writer.Write([]byte("This API expects POST requests"))
	case "OPTIONS":
		writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		writer.Write([]byte("OK"))
	case "POST":
		var body Body
		decoder := json.NewDecoder(request.Body)
		if err := decoder.Decode(&body); err != nil {
			http.Error(writer, fmt.Sprintf("Error parsing JSON %s: %s", request.Body, err),
				http.StatusBadRequest)
			return
		}

		response, err := handleBody(body, model)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Error from handleBody: %s", err),
				http.StatusBadRequest)
			return
		}

		responseBytes, err := json.Marshal(response)
		if err != nil {
			http.Error(writer, fmt.Sprintf("Error marshaling JSON %s: %s", response, err),
				http.StatusInternalServerError)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.Write(responseBytes)
	default:
		http.Error(writer, fmt.Sprintf("HTTP method not allowed"),
			http.StatusMethodNotAllowed)
		return
	}
}

func handleBody(body Body, model model.Model) (*Response, error) {
	log.Printf("-- Got body %v", body)

	if body.ResetModel {
		if err := model.Reset(); err != nil {
			return nil, fmt.Errorf("Error from model.Reset: %s", err)
		}
	}

	if body.DeviceUid == "" {
		return nil, fmt.Errorf("Blank DeviceUid")
	}
	device, err := model.FindOrCreateDeviceByUid(body.DeviceUid)
	if err != nil {
		return nil, fmt.Errorf("Error from FindOrCreateDeviceByUid: %s", err)
	}
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

	todos, err := model.ListTodos()
	if err != nil {
		return nil, fmt.Errorf("Error from ListTodos: %s", err)
	}

	response := Response{
		DeviceId:               device.Id,
		ActionToSyncIdToOutput: mapIntIntToMapStringInt(device.ActionToSyncIdToOutput),
		Todos: todos,
	}
	return &response, nil
}

// returns output -- the new TodoID if TODOS/ADD_TODOS, the number of rows updated
// for other types
func handleActionToSync(actionToSync model.ActionToSync,
	model model.Model, tempIdToId map[int]int) (int, error) {

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
		todo, err := model.CreateTodo(actionToSync)
		if err != nil {
			return 0, fmt.Errorf("Error from CreateTodo: %s", err)
		}
		return todo.Id, nil

	case "TODO/UPDATE_TODO":
		log.Printf("  Calling UpdateTodo(%v)", actionToSync)
		output, err := model.UpdateTodo(actionToSync, todoId)
		if err != nil {
			return 0, fmt.Errorf("Error from SetCompleted: %s", err)
		}
		return output, nil

	case "TODOS/DELETE_TODO":
		log.Printf("  Calling DeleteTodo(%v)", actionToSync)
		output, err := model.DeleteTodo(todoId)
		if err != nil {
			return 0, fmt.Errorf("Error from DeleteTodo: %s", err)
		}
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
