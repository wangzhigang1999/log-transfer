package main

import (
	"github.com/gorilla/websocket"
	svc "log-transfer/pkg"
	"log/slog"
	"net/http"
	"os"
)

var upgrade = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	opts := slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelInfo,
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, &opts)))
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrade.CheckOrigin = svc.OriginChecker

	// upgrade this connection to a WebSocket
	ws, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading connection", "error", err)
		return
	}
	defer ws.Close()

	// log conn details
	slog.Info("client Connected.", "remote addr", ws.RemoteAddr().String())

	target := svc.ReaderMSG{}

	err = svc.HandleRequest(ws, &target)
	if err != nil {
		return
	}

	go svc.HandleCMD(ws)

	svc.TransferLog(ws, &target)
}

func main() {

	http.HandleFunc("/ws", wsEndpoint)
	http.HandleFunc("/log", wsEndpoint)
	slog.Error("Server quit: ", http.ListenAndServe(":8080", nil))
}
