package main

import (
	"./models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
)

type Response struct {
	DeviceId               int
	ActionToSyncIdToOutput map[int]int
}

func (response *Response) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		DeviceId               int            `json:"deviceId"`
		ActionToSyncIdToOutput map[string]int `json:"actionToSyncIdToOutput"`
	}{
		DeviceId:               response.DeviceId,
		ActionToSyncIdToOutput: mapIntIntToMapStringInt(response.ActionToSyncIdToOutput),
	})
}

func mapIntIntToMapStringInt(input map[int]int) map[string]int {
	output := map[string]int{}
	for k, v := range input {
		output[strconv.Itoa(k)] = v
	}
	return output
}

func handleRequest(writer http.ResponseWriter, request *http.Request,
	model models.Model) {

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

	responseJson, err := json.Marshal(response)
	if err != nil {
		http.Error(writer, fmt.Sprintf("Error marshaling JSON %s: %s", response, err),
			http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.Write(responseJson)
}

func handleBody(body Body, model models.Model) (*Response, error) {
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
		var todoId int
		if actionToSync.Type != "TODOS/ADD_TODO" {
			if actionToSync.TodoIdMaybeTemp < 0 {
				var ok bool
				todoId, ok = tempIdToId[actionToSync.TodoIdMaybeTemp]
				if !ok {
					return nil,
						fmt.Errorf("Don't know todoId for temp id in action %v", actionToSync)
				}
			} else if actionToSync.TodoIdMaybeTemp > 0 {
				todoId = actionToSync.TodoIdMaybeTemp
			} else {
				return nil, fmt.Errorf("Invalid TodoIdMaybeTemp in action %v", actionToSync)
			}
		}

		existingOutput := device.ActionToSyncIdToOutput[actionToSync.Id]
		if existingOutput == 0 {
			var output int
			var err error
			switch actionToSync.Type {

			case "TODOS/ADD_TODO":
				log.Printf("  Calling CreateTodo(%v)", actionToSync)
				todo, err := model.CreateTodo(actionToSync)
				if err != nil {
					return nil, fmt.Errorf("Error from CreateTodo: %s", err)
				}
				tempIdToId[actionToSync.TodoIdMaybeTemp] = todo.Id
				output = todo.Id

			case "TODO/SET_COMPLETED":
				log.Printf("  Calling SetCompleted(%v)", actionToSync)
				output, err = model.SetCompleted(actionToSync.Completed, todoId)
				if err != nil {
					return nil, fmt.Errorf("Error from SetCompleted: %s", err)
				}

			default:
				return nil, fmt.Errorf("Unknown type in actionToSync: %v", actionToSync)
			}

			// immutable edit so we don't corrupt MemoryModel
			device.ActionToSyncIdToOutput = immutableSetForMapIntInt(
				device.ActionToSyncIdToOutput, actionToSync.Id, output)
		} else {
			if actionToSync.Type == "TODOS/ADD_TODO" {
				tempIdToId[actionToSync.TodoIdMaybeTemp] =
					device.ActionToSyncIdToOutput[actionToSync.Id]
			}
		}
	}
	model.UpdateDeviceActionToSyncIdToOutputJson(device)

	response := Response{
		DeviceId:               device.Id,
		ActionToSyncIdToOutput: device.ActionToSyncIdToOutput,
	}
	return &response, nil
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
