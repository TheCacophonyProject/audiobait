package main

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/TheCacophonyProject/audiobait/audiofilelibrary"
	"github.com/TheCacophonyProject/event-reporter/eventclient"
	"github.com/stretchr/testify/assert"
)

type mockSoundCard struct {
	err error
}

func (msc mockSoundCard) Play(audioFileName string, volume int) error {
	return msc.err
}

func newMockSoundCard(err error) mockSoundCard {
	return mockSoundCard{
		err: err,
	}
}

func mockOpenLibrary(filesMap map[int]string, err error) {
	openLibrary = func(soundDir string) (*audiofilelibrary.AudioFileLibrary, error) {
		return &audiofilelibrary.AudioFileLibrary{
			FilesByID: filesMap,
		}, err
	}
}

func mockSaveEvent(err error) **eventclient.Event {
	var event *eventclient.Event
	saveEvent = func(e eventclient.Event) error {
		event = &e
		return err
	}
	return &event
}

func TestPlayFromId(t *testing.T) {
	newFakeNow()
	log.Println("testing playing audio with event")
	mockOpenLibrary(map[int]string{1: "a"}, nil)
	event := mockSaveEvent(nil)
	testPlayer := player{
		soundCard: newMockSoundCard(nil),
	}
	played, err := testPlayer.PlayFromId(1, 2, 3, &eventclient.Event{})
	assert.NoError(t, err)
	assert.True(t, played)
	assert.Equal(t, eventclient.Event{
		Type:      "audioBait",
		Timestamp: now(),
		Details: map[string]interface{}{
			"fileId":   1,
			"priority": 3,
			"volume":   2,
		},
	}, **event)

	log.Println("testing played audio with no event")
	event = mockSaveEvent(nil)
	played, err = testPlayer.PlayFromId(1, 2, 3, nil)
	assert.NoError(t, err)
	assert.True(t, played)
	var expectedEvent *eventclient.Event
	assert.Equal(t, expectedEvent, *event)

	log.Println("testing failed library open")
	libraryOpenFail := errors.New("failed to open library")
	mockOpenLibrary(nil, libraryOpenFail)
	played, err = testPlayer.PlayFromId(1, 2, 3, nil)
	assert.Equal(t, libraryOpenFail, err)
	assert.False(t, played)
	assert.Equal(t, expectedEvent, *event)

	log.Println("testing failed to find file in library")
	mockOpenLibrary(map[int]string{1: "a"}, nil)
	played, err = testPlayer.PlayFromId(2, 3, 4, nil)
	assert.Error(t, err)
	assert.False(t, played)
	assert.Equal(t, expectedEvent, *event)

	log.Println("testing failed to play audio")
	soundcardError := errors.New("some soundcard error")
	testPlayer.soundCard = newMockSoundCard(soundcardError)
	played, err = testPlayer.PlayFromId(1, 2, 3, nil)
	assert.False(t, played)
	assert.Equal(t, soundcardError, err)
	assert.Equal(t, expectedEvent, *event)
}

func newFakeNow() {
	n := time.Now()
	now = func() time.Time {
		return n
	}
}
