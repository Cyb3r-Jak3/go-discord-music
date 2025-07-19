package bot

import (
	"context"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/disgolink/v3/lavalink"
)

func (b *Bot) onPlayerPause(_ disgolink.Player, event lavalink.PlayerPauseEvent) {
	b.logger.Infof("player paused, %#v", event)
}

func (b *Bot) onPlayerResume(_ disgolink.Player, event lavalink.PlayerResumeEvent) {
	b.logger.Infof("player resumed, %#v", event)
}

func (b *Bot) onTrackStart(_ disgolink.Player, event lavalink.TrackStartEvent) {
	b.logger.Infof("track started, guild: %s, track: %#v", event.GuildID(), event.Track)
}

func (b *Bot) onTrackEnd(player disgolink.Player, event lavalink.TrackEndEvent) {
	if !event.Reason.MayStartNext() {
		return
	}

	queue := b.Queues.Get(event.GuildID())
	var (
		nextTrack lavalink.Track
		ok        bool
	)
	switch queue.Type {
	case QueueTypeNormal:
		nextTrack, ok = queue.Next()

	case QueueTypeRepeatTrack:
		nextTrack = event.Track

	case QueueTypeRepeatQueue:
		queue.Add(event.Track)
		nextTrack, ok = queue.Next()
	}

	if !ok {
		return
	}
	if err := player.Update(context.TODO(), lavalink.WithTrack(nextTrack)); err != nil {
		b.logger.Errorf("error updating player track: %v", err)
	}
}

func (b *Bot) onTrackException(_ disgolink.Player, event lavalink.TrackExceptionEvent) {
	b.logger.Errorf("track exception: %#v", event)
}

func (b *Bot) onTrackStuck(_ disgolink.Player, event lavalink.TrackStuckEvent) {
	b.logger.Warnf("track stuck: %#v", event)
}

func (b *Bot) onWebSocketClosed(_ disgolink.Player, event lavalink.WebSocketClosedEvent) {
	b.logger.Warnf("websocket closed: %#v", event)
}

func (b *Bot) onUnknownEvent(_ disgolink.Player, e lavalink.UnknownEvent) {
	b.logger.Warnf("unknown event: %s, data: %s", e.Type(), string(e.Data))
}
