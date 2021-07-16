package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Tech-With-Tim/Socket-Api/utils"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/require"
)

var expectedResponses = [4]interface{}{
	pingResponse{Operation: "0"},
	redisReponse{Response: "world"},
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
		err := ws.WriteJSON(map[string]interface{}{"op": "0"})
		require.NoError(t, err)
	}
	for count := 0; count < 5; {
		var res pingResponse
		err = ws.ReadJSON(&res)
		require.NoError(t, err)
		if res.Data.Time == "" && res.Operation != "" {
			require.Contains(t, expectedResponses, res)
			count++
		}
	}
}

func recieveRedisSub(t *testing.T, wg *sync.WaitGroup, testServerUrl string) {
	defer wg.Done()
	var count int
	url := "ws" + strings.TrimPrefix(testServerUrl, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()
	for count < 10 {
		var res redisReponse
		err = ws.ReadJSON(&res)
		require.NoError(t, err)
		if res.Data.Time == "" && res.Response != "" {
			require.Contains(t, expectedResponses, res)
			count++
		}
	}
}

func publishRedis(t *testing.T, wg *sync.WaitGroup, testServerUrl string) {
	defer wg.Done()

	url := "ws" + strings.TrimPrefix(testServerUrl, "http")
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.NoError(t, err)
	defer ws.Close()

	err = ws.WriteJSON(map[string]interface{}{"op": "1",
		"d": map[string]interface{}{
			"hello": "world",
		},
	})
	require.NoError(t, err)
}

func TestSockets(t *testing.T) {
	var wg sync.WaitGroup
	//Test pings
	testServer := runTestServer(t)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go pingHandler(t, &wg, testServer.URL)
	}

	for count := 0; count < 5; count++ {
		wg.Add(1)
		go recieveRedisSub(t, &wg, testServer.URL)
	}
	time.Sleep(5 * time.Second)
	for count := 0; count < 10; count++ {
		wg.Add(1)
		time.Sleep(20 * time.Microsecond)
		go publishRedis(t, &wg, testServer.URL)
	}
	wg.Wait()
}

type data struct {
	Time string `json:"t"`
}
type pingResponse struct {
	Operation string `json:"op"`
	Data      data   `json:"d"`
}

type redisReponse struct {
	Response string `json:"hello"`
	Data     data   `json:"d"`
}
