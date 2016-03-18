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
	Id        int    `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Completed bool   `json:"completed"`
}

type Model interface {
	Reset() error
	FindOrCreateDeviceByUid(uid string) (*Device, error)
	UpdateDeviceActionToSyncIdToOutputJson(device *Device) error
	CreateTodo(action ActionToSync) (*Todo, error)
}
