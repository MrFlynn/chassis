package chassis

import (
	"errors"
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/disgoorg/disgo/discord"
)

// Error is a union of two errors. One for internal use and
// another that is to be displayed to the user.
type Error struct {
	Display  error
	Internal error
}

// Error returns the string value of the internal error value.
func (e Error) Error() (s string) {
	return e.Internal.Error()
}

var multiLineErrorMatch = regexp.MustCompile(`^[A-z]|\n\w`)

// Presentable returns the displayable error formatted with proper
// punctuation. Works with joined errors.
func (e Error) Presentable() (s string) {
	return multiLineErrorMatch.ReplaceAllStringFunc(e.Display.Error(), func(s string) string {
		var formatted string

		for len(s) > 0 {
			r, sz := utf8.DecodeRuneInString(s)

			if unicode.IsLower(r) {
				formatted += string(unicode.ToUpper(r))
			} else if r == '\n' {
				formatted += "." + string(r)
			}

			s = s[sz:]
		}

		return formatted
	}) + "."
}

func createError[M Message](err error) (message M) {
	var (
		botError Error
		value    string
	)

	if errors.As(err, &botError) {
		value = botError.Presentable()
	} else {
		value = err.Error()
	}

	switch any(message).(type) {
	case discord.MessageCreate:
		message = any(discord.NewMessageCreateBuilder().SetEmbeds(
			discord.NewEmbedBuilder().
				SetTitle(":x: Something went wrong :x:").
				SetDescription(value).
				Build(),
		).Build()).(M)
	case discord.MessageUpdate:
		message = any(discord.NewMessageUpdateBuilder().SetEmbeds(
			discord.NewEmbedBuilder().
				SetTitle(":x: Something went wrong :x:").
				SetDescription(value).
				Build(),
		).Build()).(M)
	}

	return
}
