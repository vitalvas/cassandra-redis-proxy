package app

import (
	"slices"
	"strings"

	"github.com/tidwall/redcon"
)

// Use this function to accept or deny the connection
func (app *App) redisAcceptHandler(conn redcon.Conn) bool {
	app.logger.Debug("connection accepted", "remote_addr", conn.RemoteAddr())

	// by default we accept all connections
	return true
}

// This is called when the connection has been closed
func (app *App) redisClosedHandler(conn redcon.Conn, err error) {
	if err != nil {
		app.logger.Warn("connection closed", "remote_addr", conn.RemoteAddr(), "error", err)
	} else {
		app.logger.Debug("connection closed", "remote_addr", conn.RemoteAddr())
	}
}

func (app *App) redisRequestHandler(conn redcon.Conn, cmd redcon.Command) {
	if cmd.Args == nil {
		conn.WriteError("ERR wrong number of arguments")
		return
	}

	args := bytesToStringSlice(cmd.Args)

	command := strings.ToLower(args[0])

	remote := conn.RemoteAddr()

	app.logger.Debug("redis request", "remote", remote, "command", command, "args", args)

	// skip some commands without logging
	if slices.Contains([]string{"info", "command"}, command) {
		conn.WriteError("ERR unknown command: " + command)
		return
	}

	actionFunc, ok := redisCommands[command]
	if !ok {
		app.logger.Warn("unknown command", "remote", remote, "command", command, "args", args)
		conn.WriteError("ERR unknown command: " + command)
		return
	}

	if err := actionFunc(app, conn, args); err != nil {
		app.logger.Warn("command failed", "remote", remote, "command", command, "error", err)
		conn.WriteError("ERR " + err.Error())
	}
}

func bytesToStringSlice(b [][]byte) []string {
	s := make([]string, len(b))
	for i, v := range b {
		s[i] = string(v)
	}
	return s
}
