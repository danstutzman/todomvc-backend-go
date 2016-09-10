package web

import (
	"fmt"
	models "github.com/danielstutzman/todomvc-backend-go/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

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
