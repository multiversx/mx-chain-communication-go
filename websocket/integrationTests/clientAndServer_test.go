package integrationTests

import (
	"crypto/rand"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-communication-go/testscommon"
	"github.com/multiversx/mx-chain-core-go/data/outport"
	"github.com/stretchr/testify/require"
)

func TestStartServerAddClientAndSendData(t *testing.T) {
	port := getFreePort()
	serverURL := "localhost:" + port
	wsServer, err := createServer(serverURL, &testscommon.LoggerMock{})
	require.Nil(t, err)

	wg := &sync.WaitGroup{}

	wg.Add(1)

	_ = wsServer.SetPayloadHandler(&testscommon.PayloadHandlerStub{
		ProcessPayloadCalled: func(payload []byte, topic string, version uint32) error {
			require.Equal(t, []byte("test"), payload)
			require.Equal(t, uint32(1), version)
			wg.Done()
			return nil
		},
	})

	clientURL := "ws://localhost:" + port
	wsClient, err := createClient(clientURL, &testscommon.LoggerMock{})
	require.Nil(t, err)

	for {
		err = wsClient.Send([]byte("test"), outport.TopicSaveAccounts)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	require.Nil(t, err)

	_ = wsClient.Close()
	_ = wsServer.Close()
	wg.Wait()
}

func TestStartServerAddClientAndCloseClientAndServerShouldReceiveClose(t *testing.T) {
	port := getFreePort()

	receivedClose := make(chan struct{})
	payloadWasProcessed := atomic.Bool{}
	payloadWasProcessed.Store(false)

	serverReceivedCloseMessage := atomic.Bool{}
	serverReceivedCloseMessage.Store(false)
	log := &testscommon.LoggerStub{
		InfoCalled: func(message string, args ...interface{}) {
			if strings.Contains(message, "connection closed") && payloadWasProcessed.Load() {
				serverReceivedCloseMessage.Store(true)
				receivedClose <- struct{}{}
			}
		},
	}

	serverURL := "localhost:" + port
	wsServer, err := createServer(serverURL, log)
	require.Nil(t, err)

	_ = wsServer.SetPayloadHandler(&testscommon.PayloadHandlerStub{
		ProcessPayloadCalled: func(payload []byte, topic string, version uint32) error {
			require.Equal(t, []byte("test"), payload)
			payloadWasProcessed.Store(true)
			return nil
		},
	})

	clientURL := "ws://localhost:" + port
	wsClient, err := createClient(clientURL, &testscommon.LoggerMock{})
	require.Nil(t, err)
	time.Sleep(time.Second)

	for {
		err = wsClient.Send([]byte("test"), outport.TopicSaveBlock)
		if err == nil {
			break
		}
	}

	err = wsClient.Close()
	require.Nil(t, err)
	_ = wsServer.Close()
	<-receivedClose
	require.True(t, serverReceivedCloseMessage.Load())
}

func TestStartServerStartClientCloseServer(t *testing.T) {
	port := getFreePort()
	serverURL := "localhost:" + port
	wsServer, err := createServer(serverURL, &testscommon.LoggerMock{})
	require.Nil(t, err)

	var sentMessages []string
	var receivedMessages []string

	wg := &sync.WaitGroup{}
	wg.Add(1)

	numMessagesReceived := 0
	payloadHandler := &testscommon.PayloadHandlerStub{
		ProcessPayloadCalled: func(payload []byte, topic string, version uint32) error {
			receivedMessages = append(receivedMessages, string(payload))
			numMessagesReceived++
			if numMessagesReceived == 200 {
				wg.Done()
			}
			return nil
		},
	}
	_ = wsServer.SetPayloadHandler(payloadHandler)

	clientURL := "ws://localhost:" + port
	wsClient, err := createClient(clientURL, &testscommon.LoggerMock{})
	require.Nil(t, err)

	for idx := 0; idx < 100; idx++ {
		message := fmt.Sprintf("%d", idx)
		for {
			err = wsClient.Send([]byte(message), outport.TopicSaveBlock)
			if err == nil {
				sentMessages = append(sentMessages, message)
				break
			} else {
				time.Sleep(300 * time.Millisecond)
			}
		}
	}

	err = wsServer.Close()
	require.Nil(t, err)

	time.Sleep(5 * time.Second)
	// start the server again
	wsServer, err = createServer(serverURL, &testscommon.LoggerMock{})
	_ = wsServer.SetPayloadHandler(payloadHandler)
	require.Nil(t, err)

	for idx := 100; idx < 200; idx++ {
		message := fmt.Sprintf("%d", idx)
		for {
			err = wsClient.Send([]byte(message), outport.TopicSaveBlock)
			if err == nil {
				sentMessages = append(sentMessages, message)
				break
			} else {
				time.Sleep(300 * time.Millisecond)
			}
		}
	}

	wg.Wait()
	err = wsClient.Close()
	require.Nil(t, err)
	err = wsServer.Close()
	require.Nil(t, err)

	require.Equal(t, 200, numMessagesReceived)
	require.Equal(t, sentMessages, receivedMessages)
}

func TestStartServerStartClientAndSendABigMessage(t *testing.T) {
	port := getFreePort()
	serverURL := "localhost:" + port
	wsServer, err := createServer(serverURL, &testscommon.LoggerMock{})
	require.Nil(t, err)

	wg := &sync.WaitGroup{}
	wg.Add(1)

	const thirtyMB = 30 * 1024 * 1024
	myBigMessage := generateLargeByteArray(thirtyMB)
	payloadHandler := &testscommon.PayloadHandlerStub{
		ProcessPayloadCalled: func(payload []byte, topic string, version uint32) error {
			defer wg.Done()
			require.Equal(t, myBigMessage, payload)
			require.Equal(t, outport.TopicSaveBlock, topic)
			return nil
		},
	}
	_ = wsServer.SetPayloadHandler(payloadHandler)

	clientURL := "ws://localhost:" + port
	wsClient, err := createClient(clientURL, &testscommon.LoggerMock{})
	require.Nil(t, err)

	for {
		err = wsClient.Send(myBigMessage, outport.TopicSaveBlock)
		if err == nil {
			break
		} else {
			time.Sleep(300 * time.Millisecond)
		}
	}
	wg.Wait()
	_ = wsServer.Close()
	_ = wsClient.Close()
}

func TestStartServerStartClientAndSendMultipleGoRoutines(t *testing.T) {
	port := getFreePort()
	serverURL := "localhost:" + port
	wsServer, err := createServer(serverURL, &testscommon.LoggerMock{})
	require.Nil(t, err)

	mutex := sync.RWMutex{}
	myMap := make(map[string]struct{})
	wg := sync.WaitGroup{}
	wg.Add(10000)
	for i := 0; i < 10000; i++ {
		myMap[fmt.Sprintf("%d", i)] = struct{}{}
	}

	payloadHandler := &testscommon.PayloadHandlerStub{
		ProcessPayloadCalled: func(payload []byte, topic string, version uint32) error {
			mutex.Lock()
			_, found := myMap[string(payload)]
			if !found {
				mutex.Unlock()
				return nil
			}
			delete(myMap, string(payload))
			mutex.Unlock()

			fmt.Println("received message", "payload", string(payload), "topic", topic)
			wg.Done()
			return nil
		},
	}
	_ = wsServer.SetPayloadHandler(payloadHandler)

	clientURL := "ws://localhost:" + port
	wsClient, err := createClient(clientURL, &testscommon.LoggerMock{})
	require.Nil(t, err)

	//send message to server multiple go routines
	// generate 1000 go routines, every go routine will send 10 message
	sendMultipleMessages := func(idx int) {
		for j := 0; j < 10; j++ {
			for {
				errSend := wsClient.Send([]byte(fmt.Sprintf("%d", idx*10+j)), outport.TopicSaveAccounts)
				if errSend == nil {
					break
				} else {
					time.Sleep(300 * time.Millisecond)
				}
			}

		}
	}
	for idx := 0; idx < 1000; idx++ {
		go sendMultipleMessages(idx)
	}

	wg.Wait()
	mutex.RLock()
	require.Len(t, myMap, 0)
	mutex.RUnlock()
}

func generateLargeByteArray(size int) []byte {
	bytes := make([]byte, size)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Println("failed to generate random bytes:", err)
	}
	return bytes
}
