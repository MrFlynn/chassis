package chassis

import (
	"context"
	"log/slog"
	"maps"
	"slices"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/google/go-cmp/cmp"
)

func compareHandlerMaps[M Message](x, y map[string]CommandHandler[M]) bool {
	return slices.Equal(slices.Collect(maps.Keys(x)), slices.Collect(maps.Keys(y)))
}

func compareHandlers(
	t *testing.T,
	frame *Frame,
	expectedCommandHandlers map[string]CommandHandler[discord.MessageCreate],
	expectedDeferredCommandHandlers map[string]CommandHandler[discord.MessageUpdate],
) {
	if diff := cmp.Diff(
		expectedCommandHandlers,
		frame.commandHandlers,
		cmp.Comparer(compareHandlerMaps[discord.MessageCreate]),
	); diff != "" {
		t.Errorf("Mismatch between command handlers (-want +got):\n%s", diff)
	}

	if diff := cmp.Diff(
		expectedDeferredCommandHandlers,
		frame.deferredCommandHandlers,
		cmp.Comparer(compareHandlerMaps[discord.MessageUpdate]),
	); diff != "" {
		t.Errorf("Mismatch between deferred command handlers (-want +got):\n%s", diff)
	}
}

func exampleHandler[M Message](
	_ context.Context,
	_ *events.ApplicationCommandInteractionCreate,
	_ *slog.Logger,
) (message M, err error) {
	return
}

func TestAttachHandlers(t *testing.T) {
	t.Run("single immediate handler", func(t *testing.T) {
		frame := &Frame{
			Commands: []discord.ApplicationCommandCreate{
				discord.SlashCommandCreate{Name: "test"},
			},
		}

		AttachHandlers(frame, SlashCommandHandlerRef[discord.SlashCommandCreate, discord.MessageCreate]{
			Command: PointerOf(frame.Commands[0].(discord.SlashCommandCreate)),
			Handler: exampleHandler[discord.MessageCreate],
		})

		compareHandlers(t, frame, map[string]CommandHandler[discord.MessageCreate]{
			"/test": exampleHandler[discord.MessageCreate],
		}, nil)
	})

	t.Run("single deferred handler", func(t *testing.T) {
		frame := &Frame{
			Commands: []discord.ApplicationCommandCreate{
				discord.SlashCommandCreate{Name: "test"},
			},
		}

		AttachHandlers(frame, SlashCommandHandlerRef[discord.SlashCommandCreate, discord.MessageUpdate]{
			Command: PointerOf(frame.Commands[0].(discord.SlashCommandCreate)),
			Handler: exampleHandler[discord.MessageUpdate],
		})

		compareHandlers(t, frame, nil, map[string]CommandHandler[discord.MessageUpdate]{
			"/test": exampleHandler[discord.MessageUpdate],
		})
	})

	t.Run("multi handlers", func(t *testing.T) {
		frame := &Frame{
			Commands: []discord.ApplicationCommandCreate{
				discord.SlashCommandCreate{Name: "test"},
				discord.SlashCommandCreate{Name: "help"},
			},
		}

		AttachHandlers(frame, SlashCommandHandlerRef[discord.SlashCommandCreate, discord.MessageCreate]{
			Command: PointerOf(frame.Commands[0].(discord.SlashCommandCreate)),
			Handler: exampleHandler[discord.MessageCreate],
		})

		AttachHandlers(frame, SlashCommandHandlerRef[discord.SlashCommandCreate, discord.MessageUpdate]{
			Command: PointerOf(frame.Commands[1].(discord.SlashCommandCreate)),
			Handler: exampleHandler[discord.MessageUpdate],
		})

		compareHandlers(t, frame,
			map[string]CommandHandler[discord.MessageCreate]{
				"/test": exampleHandler[discord.MessageCreate],
			},
			map[string]CommandHandler[discord.MessageUpdate]{
				"/help": exampleHandler[discord.MessageUpdate],
			},
		)
	})

	t.Run("nested handlers", func(t *testing.T) {
		frame := &Frame{
			Commands: []discord.ApplicationCommandCreate{
				discord.SlashCommandCreate{
					Name: "hello",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionSubCommand{
							Name: "world",
						},
					},
				},
			},
		}

		AttachHandlers(frame, SlashCommandHandlerRef[discord.ApplicationCommandOptionSubCommand, discord.MessageCreate]{
			Command: PointerOf(frame.Commands[0].(discord.SlashCommandCreate).
				Options[0].(discord.ApplicationCommandOptionSubCommand),
			),
			Handler: exampleHandler[discord.MessageCreate],
		})

		compareHandlers(t, frame, map[string]CommandHandler[discord.MessageCreate]{
			"/hello/world": exampleHandler[discord.MessageCreate],
		}, nil)
	})

	t.Run("nested handlers within group", func(t *testing.T) {
		frame := &Frame{
			Commands: []discord.ApplicationCommandCreate{
				discord.SlashCommandCreate{
					Name: "top",
					Options: []discord.ApplicationCommandOption{
						discord.ApplicationCommandOptionSubCommandGroup{
							Name: "middle",
							Options: []discord.ApplicationCommandOptionSubCommand{
								{Name: "bottom"},
							},
						},
					},
				},
			},
		}

		AttachHandlers(frame, SlashCommandHandlerRef[discord.ApplicationCommandOptionSubCommand, discord.MessageCreate]{
			Command: PointerOf(frame.Commands[0].(discord.SlashCommandCreate).
				Options[0].(discord.ApplicationCommandOptionSubCommandGroup).
				Options[0],
			),
			Handler: exampleHandler[discord.MessageCreate],
		})

		compareHandlers(t, frame, map[string]CommandHandler[discord.MessageCreate]{
			"/top/middle/bottom": exampleHandler[discord.MessageCreate],
		}, nil)
	})
}
