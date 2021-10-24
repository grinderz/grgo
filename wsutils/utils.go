package wsutils


import (
	"strings"

	"github.com/gorilla/websocket"
)

func Close(conn *websocket.Conn) error {
	return conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure,
			""))
}

func IsClosedWebsocketError(err error) bool {
	str := err.Error()
	if strings.Contains(str, "websocket: close sent") {
		return true
	}
	return false
}
