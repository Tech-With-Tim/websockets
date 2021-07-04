package server

import (
	"github.com/Tech-With-Tim/Socket-Api/utils"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

var expectedResponses = [4]interface{}{
	pingResponse{Operation: "0"},
	redisReponse{Response: "world"},
	pingResponse{Operation: "0", Data: Data{Time: "0"}},
	redisReponse{Response: "", Data: Data{Time: "0"}},
}

func runTestServer(t *testing.T) *httptest.Server {
	var err error
	s := CreateServer()
	s.Config, err = utils.LoadConfig("../", "test")
	require.NoError(t, err)
	err = s.RegisterRedisHandlers()
	require.NoError(t, err)
	err = s.RegisterCommands()
	require.NoError(t, err)
	go RedisHandler(s)
	testServer := httptest.NewServer(http.HandlerFunc(HandleConnections(s)))
	t.Cleanup(func() { defer testServer.Close() })
	return testServer
}

func pingHandler(t *testing.T, wg *sync.WaitGroup, testServerUrl string) {
	defer wg.Done()
	url := "ws" + strings.TrimPrefix(testServerUrl, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()

	for i := 0; i < 5; i++ {
		if err := ws.WriteJSON(map[string]interface{}{"op": "0"}); err != nil {
			require.NoError(t, err)
		}

		//expectedRes := &pingResponse{Operation: "0"}
		res := &pingResponse{}
		err := ws.ReadJSON(res)
		require.NoError(t, err)
		require.Contains(t, expectedResponses, *res)
		// require.Equal(t, expectedRes, res)
	}
}

func publishRedis(t *testing.T, wg *sync.WaitGroup, testServerUrl string) {
	defer wg.Done()
	url := "ws" + strings.TrimPrefix(testServerUrl, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()
	for i := 0; i < 5; i++ {
		err = ws.WriteJSON(map[string]interface{}{"op": "1",
			"d": map[string]interface{}{
				"hello": "world",
			},
		})
		require.NoError(t, err)

		var res redisReponse
		err = ws.ReadJSON(&res)
		require.NoError(t, err)

		require.Contains(t, expectedResponses, res)
	}
}

func TestSockets(t *testing.T) {
	var wg sync.WaitGroup
	//Test pings
	testServer := runTestServer(t)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go pingHandler(t, &wg, testServer.URL)
	}

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go publishRedis(t, &wg, testServer.URL)
	}
	wg.Wait()
}

type Data struct {
	Time string `json:"t"`
}
type pingResponse struct {
	Operation string `json:"op"`
	Data      Data   `json:"d"`
}

type redisReponse struct {
	Response string `json:"hello"`
	Data      Data   `json:"d"`
}
