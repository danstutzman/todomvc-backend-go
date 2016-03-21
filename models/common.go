package models

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
	Title           *string `json:"title",omitempty`
	Completed       *bool   `json:"completed",omitempty`
}

type Model interface {
	Reset() error
	FindOrCreateDeviceByUid(uid string) (Device, error)
	UpdateDeviceActionToSyncIdToOutputJson(device Device) error
	CreateTodo(action ActionToSync) (Todo, error)
	UpdateTodo(action ActionToSync, todoId int) (int, error)
	ListTodos() ([]Todo, error)
	DeleteTodo(todoInt int) (int, error)
}
