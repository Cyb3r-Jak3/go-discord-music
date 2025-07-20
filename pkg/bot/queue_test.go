package bot

import (
	"testing"

	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/snowflake/v2"
	"github.com/stretchr/testify/assert"
)

func Test_Queue_AddTracks_AppendsToQueue(t *testing.T) {
	queue := &Queue{}
	track1 := lavalink.Track{}
	track2 := lavalink.Track{}

	queue.Add(track1, track2)

	assert.Equal(t, 2, len(queue.Tracks))
	assert.Equal(t, track1, queue.Tracks[0])
	assert.Equal(t, track2, queue.Tracks[1])
}

func Test_Queue_Next_ReturnsFirstTrackAndRemovesIt(t *testing.T) {
	queue := &Queue{}
	track1 := lavalink.Track{}
	track2 := lavalink.Track{}
	queue.Add(track1, track2)

	track, ok := queue.Next()

	assert.True(t, ok)
	assert.Equal(t, track1, track)
	assert.Equal(t, 1, len(queue.Tracks))
	assert.Equal(t, track2, queue.Tracks[0])
}

func Test_Queue_Next_ReturnsFalseWhenEmpty(t *testing.T) {
	queue := &Queue{}

	track, ok := queue.Next()

	assert.False(t, ok)
	assert.Equal(t, lavalink.Track{}, track)
}

func Test_Queue_Skip_SkipsSpecifiedAmount(t *testing.T) {
	queue := &Queue{}
	track1 := lavalink.Track{}
	track2 := lavalink.Track{}
	track3 := lavalink.Track{}
	queue.Add(track1, track2, track3)

	track, ok := queue.Skip(2)

	assert.True(t, ok)
	assert.Equal(t, track3, track)
	assert.Equal(t, 1, len(queue.Tracks))
	assert.Equal(t, track3, queue.Tracks[0])
}

func Test_Queue_Skip_ReturnsFalseWhenEmpty(t *testing.T) {
	queue := &Queue{}

	track, ok := queue.Skip(1)

	assert.False(t, ok)
	assert.Equal(t, lavalink.Track{}, track)
}

func Test_Queue_Clear_RemovesAllTracks(t *testing.T) {
	queue := &Queue{}
	track1 := lavalink.Track{}
	track2 := lavalink.Track{}
	queue.Add(track1, track2)

	queue.Clear()

	assert.Equal(t, 0, len(queue.Tracks))
}

func Test_QueueManager_Get_ReturnsExistingQueue(t *testing.T) {
	manager := &QueueManager{queues: make(map[snowflake.ID]*Queue)}
	guildID := snowflake.ID(123)
	queue := &Queue{Type: QueueTypeNormal}
	manager.queues[guildID] = queue

	result := manager.Get(guildID)

	assert.Equal(t, queue, result)
}

func Test_QueueManager_Get_CreatesNewQueueIfNotExists(t *testing.T) {
	manager := &QueueManager{queues: make(map[snowflake.ID]*Queue)}
	guildID := snowflake.ID(123)

	result := manager.Get(guildID)

	assert.NotNil(t, result)
	assert.Equal(t, QueueTypeNormal, result.Type)
	assert.Equal(t, 0, len(result.Tracks))
}

func Test_QueueManager_Delete_RemovesQueue(t *testing.T) {
	manager := &QueueManager{queues: make(map[snowflake.ID]*Queue)}
	guildID := snowflake.ID(123)
	manager.queues[guildID] = &Queue{}

	manager.Delete(guildID)

	_, exists := manager.queues[guildID]
	assert.False(t, exists)
}
