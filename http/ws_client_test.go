package client

import (
	"github.com/stretchr/testify/require"
	"testing"
	"context"
	"github.com/stretchr/testify/assert"
)

func TestWSClient_ShouldNotSendMessageBeforeReady(t *testing.T) {
	defer startTestServer().Shutdown(context.TODO())

	uri := "ws://localhost:9000/ws"
	for i := 0; i < 10; i++ {
		cli, err := NewWSClient(uri)
		require.NoError(t, err)
		cli.WriteMessage(nil)
		msg, err := cli.ReadMessage()
		require.NoError(t, err)
		require.Equal(t, 0, len(msg))
	}
}

func TestWSClient_ConnectAndDisconnect(t *testing.T) {
	defer startTestServer().Shutdown(context.TODO())

	uri := "ws://localhost:9000/ws"
	for i := 0; i < 10; i++ {
		cli, err := NewWSClient(uri)
		require.NoError(t, err)
		cli.WriteMessage(nil)
		msg, err := cli.ReadMessage()
		require.NoError(t, err)
		require.Equal(t, 0, len(msg))
		assert.NoError(t, cli.Close())
	}
}

