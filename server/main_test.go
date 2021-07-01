package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

var s *Server
var testServer *httptest.Server

func TestMain(m *testing.M) {
	s = CreateServer()
	//http.HandleFunc("/", HandleConnections(s))
	testServer = httptest.NewServer(http.HandlerFunc(HandleConnections(s)))
	defer testServer.Close()
	go HandleChallenges(s)
	os.Exit(m.Run())
}

func pingHandler(t *testing.T){
	url := "ws" + strings.TrimPrefix(testServer.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()

	for i:=0; i<5; i++ {
		if err := ws.WriteJSON(map[string]interface{}{"op":"0"}); err != nil {
			require.NoError(t, err)
		}
		expectedRes := &pingResponse{operation: 0}
		fmt.Println(expectedRes)
		res := &pingResponse{}
		err := ws.ReadJSON(res)
		require.NoError(t, err)
		require.Equal(t, expectedRes, res)
	}
}
func TestPing(t *testing.T) {
	for i:=0; i<5; i++{
		go pingHandler(t)
	}
}

type pingResponse struct {
	operation int `json:"op"`
}