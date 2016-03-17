package models

type Device struct {
	Id                     int
	Uid                    string
	ActionToSyncIdToOutput map[int]int
}

type Model interface {
	Reset() error
	FindOrCreateDeviceByUid(uid string) (*Device, error)
}
