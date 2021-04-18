package audiobaitclient

import (
	"errors"
	"testing"

	"github.com/TheCacophonyProject/event-reporter/eventclient"
	"github.com/stretchr/testify/assert"
)

func mockDBusCall(i []interface{}, err error) func(string, ...interface{}) ([]interface{}, error) {
	return func(string, ...interface{}) ([]interface{}, error) {
		return i, err
	}
}

func TestPlayedAudio(t *testing.T) {
	dbusCall = mockDBusCall([]interface{}{true}, nil)
	played, err := PlayFromId(1, 2, 3, &eventclient.Event{})
	assert.True(t, played)
	assert.NoError(t, err)
}

func TestDidNotPlayAudio(t *testing.T) {
	dbusCall = mockDBusCall([]interface{}{false}, nil)
	played, err := PlayFromId(1, 2, 3, nil)
	assert.False(t, played)
	assert.NoError(t, err)
}

func TestPlayError(t *testing.T) {
	dbusCall = mockDBusCall([]interface{}{true}, errors.New("an error"))
	played, err := PlayFromId(1, 1, 1, &eventclient.Event{})
	assert.False(t, played)
	assert.Error(t, err)
}

func TestPlayBadDBusReturns(t *testing.T) {
	dbusCall = mockDBusCall([]interface{}{1}, nil) // Returning wrong type
	success, err := PlayFromId(1, 1, 1, &eventclient.Event{})
	assert.False(t, success)
	assert.Error(t, err)

	dbusCall = mockDBusCall([]interface{}{true, true}, nil) // Returning too many
	success, err = PlayFromId(1, 1, 1, &eventclient.Event{})
	assert.False(t, success)
	assert.Error(t, err)

	dbusCall = mockDBusCall([]interface{}{}, nil) // Returning not enough
	success, err = PlayFromId(1, 1, 1, &eventclient.Event{})
	assert.False(t, success)
	assert.Error(t, err)
}
