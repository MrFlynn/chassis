package chassis

import (
	"context"
	"errors"
	"iter"
	"log/slog"
	"slices"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/google/go-cmp/cmp"
)

// Message is a type constraint for different message types that can be returned
// by a CommandHandler function.
type Message interface {
	discord.MessageCreate | discord.MessageUpdate
}

// CommandHandler is a function signature for a function used to respond to an
// invocation of a Discord slash command. A handler that returns the type
// discord.MessageUpdate must call event.DeferMessageCreate otherwise sending
// the response will always fail.
type CommandHandler[M Message] func(
	ctx context.Context,
	event *events.ApplicationCommandInteractionCreate,
	logger *slog.Logger,
) (message M, err error)

func executeHandler[M Message](
	ctx context.Context,
	handler CommandHandler[M],
	event *events.ApplicationCommandInteractionCreate,
	logger *slog.Logger,
) {
	message, err := handler(ctx, event, logger)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			logger.ErrorContext(ctx, "Timeout exceeded while processing command")
			return
		}

		message = createError[M](err)
	}

	switch m := any(message).(type) {
	case discord.MessageCreate:
		if err := event.CreateMessage(m, rest.WithCtx(ctx)); err != nil {
			logger.ErrorContext(ctx, "Create message failed", "error", err)
		}
	case discord.MessageUpdate:
		if _, err := event.Client().Rest().UpdateInteractionResponse(
			event.ApplicationID(), event.Token(), m, rest.WithCtx(ctx),
		); err != nil {
			logger.ErrorContext(ctx, "Update interaction failed", "error", err)
		}
	}
}

// Command is a type constraint for different command types that we can attach
// a handler to.
type Command interface {
	discord.SlashCommandCreate | discord.ApplicationCommandOptionSubCommand
}

// SlashCommandHandlerRef points to a command somewhere in the command trie
// and attaches a handler function to it.
type SlashCommandHandlerRef[C Command, M Message] struct {
	Command *C
	Handler CommandHandler[M]
}

type stackEntry struct {
	item      any
	pathSteps []string
}

func (s stackEntry) path() (p string) {
	return "/" + strings.Join(s.pathSteps, "/")
}

func findCommandPaths[C Command, M Message](
	commands []discord.ApplicationCommandCreate,
	refs ...SlashCommandHandlerRef[C, M],
) iter.Seq2[string, CommandHandler[M]] {
	stack := make([]stackEntry, 0, len(commands))
	for _, command := range commands {
		stack = append(stack, stackEntry{
			item:      command,
			pathSteps: []string{command.CommandName()},
		})
	}

	return func(yield func(string, CommandHandler[M]) bool) {
		for len(stack) > 0 {
			entry := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			if i := slices.IndexFunc(refs, func(r SlashCommandHandlerRef[C, M]) bool {
				return cmp.Equal(*r.Command, entry.item)
			}); i != -1 {
				if !yield(entry.path(), refs[i].Handler) {
					return
				}
			}

			switch command := entry.item.(type) {
			case discord.SlashCommandCreate:
				for _, opt := range command.Options {
					stack = append(stack, stackEntry{
						item:      opt,
						pathSteps: append(entry.pathSteps, opt.OptionName()),
					})
				}
			case discord.SlashCommandOptionSubCommandGroup:
				for _, opt := range command.Options {
					stack = append(stack, stackEntry{
						item:      opt,
						pathSteps: append(entry.pathSteps, opt.Name),
					})
				}
			}
		}
	}
}
