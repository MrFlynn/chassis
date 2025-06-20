package chassis

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

// Frame is a Discord bot implementation that is designed for ease of setup
// and expansion.
type Frame struct {
	Client         bot.Client
	Commands       []discord.ApplicationCommandCreate
	EventListeners events.ListenerAdapter

	// Top-level context of the running program.
	globalCtx context.Context

	// Handlers for different message response types. One for "regular" responses
	// (those responded to in 3 or fewer seconds) and deferred responses.
	commandHandlers         map[string]CommandHandler[discord.MessageCreate]
	deferredCommandHandlers map[string]CommandHandler[discord.MessageUpdate]
}

// Start configures the bot with the given key and configuration options, connects it
// to the Discord API gateway, and configures provided slash commands.
//
// If an API key is not provided when calling this function, it will be read from the
// DISCORD_TOKEN environment variable.
//
// By default, the standard gateway configuration is used and a slash command event
// listener is configured based on the attached handlers.
//
// For testing purposes, you can set the TEST_GUILD_ID environment variable with the ID
// of a Discord guild (aka server), and the slash commands will only be propogated to that
// specific guild.
func (f *Frame) Start(ctx context.Context, key string, opts ...bot.ConfigOpt) (err error) {
	f.globalCtx = ctx
	f.EventListeners.OnApplicationCommandInteraction = f.slashCommandEventListener

	f.Client, err = disgo.New(
		cmp.Or(key, os.Getenv("DISCORD_TOKEN")),
		append([]bot.ConfigOpt{
			bot.WithDefaultGateway(),
			bot.WithEventListeners(&f.EventListeners),
		}, opts...)...,
	)

	if err != nil {
		return fmt.Errorf("failed to initialize bot: %w", err)
	}

	if err := f.Client.OpenGateway(ctx); err != nil {
		return fmt.Errorf("unable to connect to gateway: %w", err)
	}

	if guildID := snowflake.GetEnv("TEST_GUILD_ID"); guildID != snowflake.ID(0) {
		if _, err := f.Client.Rest().SetGuildCommands(
			f.Client.ApplicationID(), guildID, f.Commands,
		); err != nil {
			return fmt.Errorf("unable to set guild commands: %w", err)
		}
	} else {
		if _, err := f.Client.Rest().SetGlobalCommands(
			f.Client.ApplicationID(), f.Commands,
		); err != nil {
			return fmt.Errorf("unable to set global commands: %w", err)
		}
	}

	return
}

// AttachHandlers attaches the given command handlers to the given BotBuilder instance
// so they can be used in the command event listener.
func AttachHandlers[C Command, M Message](frame *Frame, refs ...SlashCommandHandlerRef[C, M]) {
	for path, handler := range findCommandPaths(frame.Commands, refs...) {
		switch fn := any(handler).(type) {
		case CommandHandler[discord.MessageCreate]:
			if frame.commandHandlers == nil {
				frame.commandHandlers = make(map[string]CommandHandler[discord.MessageCreate])
			}

			frame.commandHandlers[path] = fn
		case CommandHandler[discord.MessageUpdate]:
			if frame.deferredCommandHandlers == nil {
				frame.deferredCommandHandlers = make(map[string]CommandHandler[discord.MessageUpdate])
			}

			frame.deferredCommandHandlers[path] = fn
		}
	}
}

func (f *Frame) slashCommandEventListener(event *events.ApplicationCommandInteractionCreate) {
	var (
		command = event.SlashCommandInteractionData().CommandPath()
		logger  = slog.With(
			slog.String("command", command),
		)
	)

	if logger.Enabled(f.globalCtx, slog.LevelDebug) {
		logger = logger.With(
			slog.Any("guild", *event.SlashCommandInteractionData().GuildID()),
		)
	}

	if handler, ok := f.deferredCommandHandlers[command]; ok && handler != nil {
		// For deferred interactions we have up to 15 minutes to respond.
		ctx, cancel := context.WithTimeout(f.globalCtx, 15*time.Minute)
		defer cancel()

		executeHandler(ctx, handler, event, logger)
		return
	}

	// cmp.Or doesn't work here because it only works on comparable types, of which
	// functions are not.
	if handler := or(
		f.commandHandlers[command],
		f.commandHandlers["/help"],
		commandNotFound,
	); handler != nil {
		// For regular interactions we have 3 seconds to respond.
		ctx, cancel := context.WithTimeout(f.globalCtx, 3*time.Second)
		defer cancel()

		executeHandler(ctx, handler, event, logger)
		return
	}
}
