package model

type MemoryModel struct {
	Devices      []Device
	NextDeviceId int
	Todos        []Todo
	NextTodoId   int
}

func NewMemoryModel() *MemoryModel {
	model := MemoryModel{}
	model.Reset()
	return &model
}

func (model *MemoryModel) Reset() {
	model.Devices = []Device{}
	model.NextDeviceId = 1
	model.Todos = []Todo{}
	model.NextTodoId = 1
}

func (model *MemoryModel) FindOrCreateDeviceByUid(uid string) Device {
	for _, device := range model.Devices {
		if device.Uid == uid {
			return device
		}
	}

	newDevice := Device{
		Id:  model.NextDeviceId,
		Uid: uid,
		ActionToSyncIdToOutput: map[int]int{},
	}
	model.Devices = append(model.Devices, newDevice)
	model.NextDeviceId += 1
	return newDevice
}

func (model *MemoryModel) CreateTodo(action ActionToSync) Todo {
	newTodo := Todo{
		Id:        model.NextTodoId,
		Title:     *action.Title,
		Completed: *action.Completed,
	}
	model.Todos = append(model.Todos, newTodo)
	model.NextTodoId += 1
	return newTodo
}

func (model *MemoryModel) UpdateDeviceActionToSyncIdToOutputJson(
	updatedDevice Device) {
	for i, device := range model.Devices {
		if device.Uid == updatedDevice.Uid {
			device.ActionToSyncIdToOutput = updatedDevice.ActionToSyncIdToOutput
			model.Devices[i] = device
		}
	}
}

func (model *MemoryModel) UpdateTodo(action ActionToSync, todoId int) int {
	for i, todo := range model.Todos {
		if todo.Id == todoId {
			if action.Completed != nil {
				todo.Completed = *action.Completed
			}
			if action.Title != nil {
				todo.Title = *action.Title
			}
			model.Todos[i] = todo
			return 1
		}
	}
	return 0
}

func (model *MemoryModel) ListTodos() []Todo {
	todosCopy := make([]Todo, len(model.Todos))
	copy(todosCopy, model.Todos)
	return todosCopy
}

func (model *MemoryModel) DeleteTodo(todoId int) int {
	numRowsDeleted := 0
	newTodos := []Todo{}
	for _, todo := range model.Todos {
		if todo.Id == todoId {
			numRowsDeleted += 1
		} else {
			newTodos = append(newTodos, todo)
		}
	}
	model.Todos = newTodos
	return numRowsDeleted
}
