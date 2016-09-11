package handlers

import (
	"fmt"
	"github.com/danielstutzman/todomvc-backend-go/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }

func TestHandleBodyNoDeviceUid(t *testing.T) {
	model := &models.MemoryModel{
		NextDeviceId: 1,
		Devices:      []models.Device{},
	}
	_, err := HandleBody(Body{}, model)
	assert.Equal(t, fmt.Errorf("Blank DeviceUid"), err)
}

func TestHandleBodyNewDevice(t *testing.T) {
	model := &models.MemoryModel{
		NextDeviceId: 2,
		Devices: []models.Device{
			{Id: 1, Uid: "earlier", ActionToSyncIdToOutput: map[int]int{}},
		},
	}
	_, err := HandleBody(Body{DeviceUid: "new"}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, []models.Device{
		{Id: 1, Uid: "earlier", ActionToSyncIdToOutput: map[int]int{}},
		{Id: 2, Uid: "new", ActionToSyncIdToOutput: map[int]int{}},
	}, model.Devices)
}

func TestHandleBodyExistingDevice(t *testing.T) {
	model := &models.MemoryModel{
		NextDeviceId: 2,
		Devices: []models.Device{
			{Id: 1, Uid: "here", ActionToSyncIdToOutput: map[int]int{}},
		},
	}
	_, err := HandleBody(Body{DeviceUid: "here"}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, []models.Device{
		{Id: 1, Uid: "here", ActionToSyncIdToOutput: map[int]int{}},
	}, model.Devices)
}

func TestCreateNewDevice(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid:     "A",
		ActionsToSync: []models.ActionToSync{},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, []models.Device{{
		Id:  1,
		Uid: "A",
		ActionToSyncIdToOutput: map[int]int{},
	}}, model.Devices)
}

func TestCreateSameDeviceTwice(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid:     "D",
		ActionsToSync: []models.ActionToSync{},
	}, model)
	assert.Equal(t, nil, err)
	_, err = HandleBody(Body{
		DeviceUid:     "D",
		ActionsToSync: []models.ActionToSync{},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, []models.Device{{
		Id:  1,
		Uid: "D",
		ActionToSyncIdToOutput: map[int]int{},
	}}, model.Devices)
}

func TestCreate2NewDevices(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid:     "B",
		ActionsToSync: []models.ActionToSync{},
	}, model)
	assert.Equal(t, nil, err)
	_, err = HandleBody(Body{
		DeviceUid:     "C",
		ActionsToSync: []models.ActionToSync{},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, []models.Device{
		{Id: 1, Uid: "B", ActionToSyncIdToOutput: map[int]int{}},
		{Id: 2, Uid: "C", ActionToSyncIdToOutput: map[int]int{}},
	}, model.Devices)
}

func TestCreateTodo(t *testing.T) {
	model := &models.MemoryModel{
		NextDeviceId: 2,
		Devices: []models.Device{
			{Id: 1, Uid: "here", ActionToSyncIdToOutput: map[int]int{}},
		},
		NextTodoId: 1,
	}
	_, err := HandleBody(Body{
		DeviceUid: "here",
		ActionsToSync: []models.ActionToSync{{
			Id:              1,
			Type:            "TODOS/ADD_TODO",
			TodoIdMaybeTemp: -1,
			Title:           stringPtr("title"),
			Completed:       boolPtr(true),
		}},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, []models.Todo{{
		Id:        1,
		Title:     "title",
		Completed: true,
	}}, model.Todos)
	assert.Equal(t, 2, model.NextTodoId)
}

func TestCreateSameTodoTwice(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid: "here",
		ActionsToSync: []models.ActionToSync{
			{
				Id:              1,
				Type:            "TODOS/ADD_TODO",
				TodoIdMaybeTemp: -1,
				Title:           stringPtr("title1"),
				Completed:       boolPtr(true),
			}, {
				Id:              1,
				Type:            "TODOS/ADD_TODO",
				TodoIdMaybeTemp: -1,
				Title:           stringPtr("title1"),
				Completed:       boolPtr(true),
			},
		},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, map[int]int{1: 1}, model.Devices[0].ActionToSyncIdToOutput)
	assert.Equal(t, []models.Todo{{
		Id:        1,
		Title:     "title1",
		Completed: true,
	}}, model.Todos)
}

func TestCreate1ThenCreate1Update2(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{{
			Id:              1,
			Type:            "TODOS/ADD_TODO",
			TodoIdMaybeTemp: -1,
			Title:           stringPtr("title1"),
			Completed:       boolPtr(false),
		}},
	}, model)
	assert.Equal(t, nil, err)
	_, err = HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{
			{
				Id:              1,
				Type:            "TODOS/ADD_TODO",
				TodoIdMaybeTemp: -1,
				Title:           stringPtr("title1"),
				Completed:       boolPtr(false),
			}, {
				Id:              2,
				Type:            "TODO/UPDATE_TODO",
				TodoIdMaybeTemp: -1,
				Completed:       boolPtr(true),
			},
		},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, map[int]int{1: 1, 2: 1}, model.Devices[0].ActionToSyncIdToOutput)
	assert.Equal(t, []models.Todo{{
		Id:        1,
		Title:     "title1",
		Completed: true,
	}}, model.Todos)
}

func TestCreate1ThenDelete1(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{{
			Id:              1,
			Type:            "TODOS/ADD_TODO",
			TodoIdMaybeTemp: -1,
			Title:           stringPtr("title"),
			Completed:       boolPtr(false),
		}},
	}, model)
	assert.Equal(t, nil, err)
	_, err = HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{{
			Id:              2,
			Type:            "TODOS/DELETE_TODO",
			TodoIdMaybeTemp: 1,
		}},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, map[int]int{1: 1, 2: 1}, model.Devices[0].ActionToSyncIdToOutput)
	assert.Equal(t, []models.Todo{}, model.Todos)
}

func TestCreate1ThenUpdate1(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{{
			Id:              1,
			Type:            "TODOS/ADD_TODO",
			TodoIdMaybeTemp: -1,
			Title:           stringPtr("title"),
			Completed:       boolPtr(false),
		}},
	}, model)
	assert.Equal(t, nil, err)
	_, err = HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{{
			Id:              2,
			Type:            "TODO/UPDATE_TODO",
			TodoIdMaybeTemp: 1,
			Completed:       boolPtr(true),
		}},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, map[int]int{1: 1, 2: 1}, model.Devices[0].ActionToSyncIdToOutput)
	assert.Equal(t, []models.Todo{{
		Id:        1,
		Title:     "title",
		Completed: true,
	}}, model.Todos)
}

func TestCreate1Delete1(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{
			{
				Id:              1,
				Type:            "TODOS/ADD_TODO",
				TodoIdMaybeTemp: -1,
				Title:           stringPtr("title"),
				Completed:       boolPtr(false),
			},
			{
				Id:              2,
				Type:            "TODOS/DELETE_TODO",
				TodoIdMaybeTemp: -1,
			},
		},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, map[int]int{1: 1, 2: 1}, model.Devices[0].ActionToSyncIdToOutput)
	assert.Equal(t, []models.Todo{}, model.Todos)
}

func TestCreate1Update1(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{
			{
				Id:              1,
				Type:            "TODOS/ADD_TODO",
				TodoIdMaybeTemp: -1,
				Title:           stringPtr("title"),
				Completed:       boolPtr(false),
			},
			{
				Id:              2,
				Type:            "TODO/UPDATE_TODO",
				TodoIdMaybeTemp: -1,
				Completed:       boolPtr(true),
			},
		},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, map[int]int{1: 1, 2: 1}, model.Devices[0].ActionToSyncIdToOutput)
	assert.Equal(t, []models.Todo{{
		Id:        1,
		Title:     "title",
		Completed: true,
	}}, model.Todos)
}

func TestCreate2Todos(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{{
			Id:              11,
			Type:            "TODOS/ADD_TODO",
			TodoIdMaybeTemp: -21,
			Title:           stringPtr("new title"),
			Completed:       boolPtr(false),
		}},
	}, model)
	assert.Equal(t, nil, err)
	_, err = HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{{
			Id:              12,
			Type:            "TODOS/ADD_TODO",
			TodoIdMaybeTemp: -22,
			Title:           stringPtr("new title 2"),
			Completed:       boolPtr(false),
		}},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, map[int]int{11: 1, 12: 2}, model.Devices[0].ActionToSyncIdToOutput)
	assert.Equal(t, []models.Todo{
		{
			Id:        1,
			Title:     "new title",
			Completed: false,
		}, {
			Id:        2,
			Title:     "new title 2",
			Completed: false,
		},
	}, model.Todos)
}

func TestCreate1Todo(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{{
			Id:              1,
			Type:            "TODOS/ADD_TODO",
			TodoIdMaybeTemp: -1,
			Title:           stringPtr("new title"),
			Completed:       boolPtr(false),
		}},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, map[int]int{1: 1}, model.Devices[0].ActionToSyncIdToOutput)
	assert.Equal(t, []models.Todo{{
		Id:        1,
		Title:     "new title",
		Completed: false,
	}}, model.Todos)
}

func TestCreateTodoUpdateTitle(t *testing.T) {
	model := models.NewMemoryModel()
	_, err := HandleBody(Body{
		DeviceUid: "A",
		ActionsToSync: []models.ActionToSync{{
			Id:              1,
			Type:            "TODOS/ADD_TODO",
			TodoIdMaybeTemp: -1,
			Title:           stringPtr("title"),
			Completed:       boolPtr(false),
		}, {
			Id:              2,
			Type:            "TODO/UPDATE_TODO",
			TodoIdMaybeTemp: -1,
			Title:           stringPtr("new title"),
		}},
	}, model)
	assert.Equal(t, nil, err)
	assert.Equal(t, []models.Todo{{
		Id:        1,
		Title:     "new title",
		Completed: false,
	}}, model.Todos)
}
