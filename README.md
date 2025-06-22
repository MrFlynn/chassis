# Chassis
[![Go Reference](https://pkg.go.dev/badge/github.com/mrflynn/chassis.svg)](https://pkg.go.dev/github.com/mrflynn/chassis)
[![Tests](https://github.com/MrFlynn/chassis/actions/workflows/test.yml/badge.svg)](https://github.com/MrFlynn/chassis/actions/workflows/test.yml)

Chassis is a library that eliminates a bunch of the boilerplate when building
new Discord bots. This library builds on top of the
[`disgo`](https://github.com/disgoorg/disgo) library for the Discord API and
adds convenience features related to:
* Bot initialization (setup and starting).
* Slash command handling.
* Error handling.

> [!NOTE]
> This library should **not** be considered stable. There may be breaking changes between
> versions. See the [Status and Contributions](#status-and-contributions) section below
> for more information.

## Features
The aim of this library is to allow implementer to skip the process of writing code
to handle setup+connecting to the API, writing event handlers for directing
slash commands to the right place, and error handling. When all of the features
are used in conjunction, an implementer can avoid doing those things and focus on writing
the actual specific logic of their bot. For more detail on specific features, see each
section below for more information.

<details>
<summary><strong>Slash Command Handlers</strong></summary>

This library provides convenience functions to associate functions with slash
commands. Once the implementer has defined their commands and handler functions
for each, all they need to do is call `AttachHandlers` on each pair of handler
functions and pointers to the specific command struct object and the library
takes care of the rest.

Internally, calls to `AttachHandlers` searches the command tree for the command
object and builds a map of each command path to each handler function.

Handlers support both regular and deferred command responses in the form of
created or updated messages. For example, if you have a long running command,
all you need to do is defer the event in your handler and return a message
update and the library will handle the rest.
</details>

<details>
<summary><strong>Error Handling</strong></summary>
If an error occurs in a handler, this library will automatic generate error
messages and present them to the user. Furthermore, if you use the provided
`Error` type, you can annotate errors for internal logging and presentation
to users of your bot. The library also formats errors with proper punctuation
and capitalization so they look nice and neat.
</details>

<details>
<summary><strong>Easy Bot Initialization</strong></summary>
Finally, this library provides a simple way of connecting your bot to the API
to start interacting with users. It handles the process of setting up the API
client, connecting it to the gateway, and adding your slash commands.
</details>

## Usage
Below you can find a very simple that demonstrates how to use this library. This
bot has two separate slash commands with separate handlers for each. It shows
how to attach handlers and how to connect the bot so it can start serving requests.

```go
package main

import (
 	"context"
 	"log/slog"
 	"os"
 	"os/signal"

 	"github.com/disgoorg/disgo/discord"
 	"github.com/disgoorg/disgo/events"
 	"github.com/mrflynn/chassis"
)

func helloHandler(
 	_ context.Context,
 	_ *events.ApplicationCommandInteractionCreate,
 	_ *slog.Logger,
) (message discord.MessageCreate, err error) {
 	return discord.NewMessageCreateBuilder().
		SetContent("world!").
		Build(), nil
}

func helpHandler(
 	_ context.Context,
 	_ *events.ApplicationCommandInteractionCreate,
 	_ *slog.Logger,
) (message discord.MessageCreate, err error) {
 	return discord.NewMessageCreateBuilder().
		SetContent("No help for you").
		Build(), nil
}

var commands = []discord.ApplicationCommandCreate{
 	discord.SlashCommandCreate{Name: "hello"},
 	discord.SlashCommandCreate{Name: "help"},
}

func main() {
 	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
 	defer cancel()

 	bot := &chassis.Frame{Commands: commands}

 	chassis.AttachHandlers(
		bot,
		chassis.SlashCommandHandlerRef[discord.SlashCommandCreate, discord.MessageCreate]{
 			Command: chassis.PointerOf(commands[0].(discord.SlashCommandCreate)),
 			Handler: helloHandler,
		},
		chassis.SlashCommandHandlerRef[discord.SlashCommandCreate, discord.MessageCreate]{
 			Command: chassis.PointerOf(commands[1].(discord.SlashCommandCreate)),
 			Handler: helpHandler,
		},
 	)

 	if err := bot.Start(ctx, ""); err != nil {
		panic(err)
 	}

 	<-ctx.Done()
}
```

If you need more complex functionality, like handlers that are methods of the
bot itself (to access custom struct fields), you can embed `Frame` and call
`AttachHandlers` on those methods.

## Motivation
As I've written a fair number of custom Discord bots for communities that I'm
a member of, I find myself writing the same boilerplate code for every bot.
This library distills and generalizes the setup, command, and error handling
code I've implemented for each so I can avoid writing the same code every time
I create a new bot. To be clear, this library is based on my tastes and needs so
it's probably perfectly generalized nor comprehensive (if you need that, just
use the underlying [`disgo`](https://github.com/disgoorg/disgo) package), but
it covers enough of the base functionality one would want in a bot (even
moderately complex ones), that I figured I would share what I built with the
world.

## Status and Contributions
This library should not be considered stable. There will probably be breaking
changes between minor versions until `1.0.0`. As I add more features there may be
a need to refactor or rewrite large portions of this library to accommodate them.
Please keep this in mind if you choose to use this library.

I probably won't accept feature requests that go beyond the scope of what's
here unless it's something I need or can see myself needing in the future.
Bug fixes are of course welcome.
