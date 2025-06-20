package chassis

import (
	"context"
	"log/slog"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

// PointerOf returns a pointer to the given value.
func PointerOf[T any](v T) (p *T) {
	return &v
}

func or[M Message](fns ...CommandHandler[M]) (h CommandHandler[M]) {
	for _, fn := range fns {
		if fn != nil {
			return fn
		}
	}

	return
}

func commandNotFound(
	ctx context.Context,
	event *events.ApplicationCommandInteractionCreate,
	logger *slog.Logger,
) (message discord.MessageCreate, err error) {
	logger.WarnContext(ctx, "Got interaction with unknown or unimplemented command.")

	return discord.NewMessageCreateBuilder().SetContentf(
		"Unknown or unimplmented command: `/%s`.", strings.Join(
			strings.Split(
				strings.TrimLeft(
					event.SlashCommandInteractionData().CommandPath(), "/",
				), "/",
			), " ",
		),
	).Build(), nil
}
