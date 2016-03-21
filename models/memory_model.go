package models

type MemoryModel struct {
	Devices      []Device
	NextDeviceId int
	Todos        []Todo
	NextTodoId   int
}

func (model *MemoryModel) Reset() error {
	model.Devices = []Device{}
	model.NextDeviceId = 1
	model.Todos = []Todo{}
	model.NextTodoId = 1
	return nil
}

func (model *MemoryModel) FindOrCreateDeviceByUid(uid string) (Device, error) {
	for _, device := range model.Devices {
		if device.Uid == uid {
			return device, nil
		}
	}

	newDevice := Device{
		Id:  model.NextDeviceId,
		Uid: uid,
		ActionToSyncIdToOutput: map[int]int{},
	}
	model.Devices = append(model.Devices, newDevice)
	model.NextDeviceId += 1
	return newDevice, nil
}

func (model *MemoryModel) CreateTodo(action ActionToSync) (Todo, error) {
	newTodo := Todo{
		Id:        model.NextTodoId,
		Title:     action.Title,
		Completed: action.Completed,
	}
	model.Todos = append(model.Todos, newTodo)
	model.NextTodoId += 1
	return newTodo, nil
}

func (model *MemoryModel) UpdateDeviceActionToSyncIdToOutputJson(
	updatedDevice Device) error {
	for i, device := range model.Devices {
		if device.Uid == updatedDevice.Uid {
			device.ActionToSyncIdToOutput = updatedDevice.ActionToSyncIdToOutput
			model.Devices[i] = device
		}
	}
	return nil
}

func (model *MemoryModel) SetCompleted(completed bool, todoId int) (int, error) {
	for i, todo := range model.Todos {
		if todo.Id == todoId {
			todo.Completed = completed
			model.Todos[i] = todo
			return 1, nil
		}
	}
	return 0, nil
}

func (model *MemoryModel) ListTodos() ([]Todo, error) {
	todosCopy := make([]Todo, len(model.Todos))
	copy(todosCopy, model.Todos)
	return todosCopy, nil
}
