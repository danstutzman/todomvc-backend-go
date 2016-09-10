package model

type Device struct {
	Id                     int
	Uid                    string
	ActionToSyncIdToOutput map[int]int
}

type Todo struct {
	Id        int    `json:"id"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type ActionToSync struct {
	Id              int     `json:"id"`
	Type            string  `json:"type"`
	TodoIdMaybeTemp int     `json:"todoIdMaybeTemp"`
	Title           *string `json:"title,omitempty"`
	Completed       *bool   `json:"completed,omitempty"`
}

type Model interface {
	Reset()
	FindOrCreateDeviceByUid(uid string) Device
	UpdateDeviceActionToSyncIdToOutputJson(device Device)
	CreateTodo(action ActionToSync) Todo
	UpdateTodo(action ActionToSync, todoId int) int
	ListTodos() []Todo
	DeleteTodo(todoInt int) int
}
