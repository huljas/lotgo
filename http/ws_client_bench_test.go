package client

import (
	"testing"
	"context"
)

//   200000	     32531 ns/op	    1271 B/op	       9 allocs/op
func BenchmarkWS_WSClient_WriteRead_EmptyMessage(b *testing.B) {
	defer startTestServer().Shutdown(context.TODO())
	uri := "ws://localhost:9000/ws"
	cli, err := NewWSClient(uri)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		err := cli.WriteMessage(nil)
		if err != nil {
			b.Error(err)
		}
		msg, err := cli.ReadMessage()
		if err != nil {
			b.Error(err)
		}
		if len(msg) > 0 {
			b.Error("Should have empty message")
		}
	}
}

//    30000	    182195 ns/op	  127242 B/op	      20 allocs/op
func BenchmarkWS_WSClient_WriteRead_BigMessage(b *testing.B) {
	defer startTestServer().Shutdown(context.TODO())
	uri := "ws://localhost:9000/ws"
	bigmessage := make([]byte, 20*1024)
	cli, err := NewWSClient(uri)
	if err != nil {
		b.Fatal(err)
	}
	for i := 0; i < b.N; i++ {
		err := cli.WriteMessage(bigmessage)
		if err != nil {
			b.Error(err)
		}
		msg, err := cli.ReadMessage()
		if err != nil {
			b.Error(err)
		}
		if len(msg) != 20*1024 {
			b.Error("Should have big message")
		}
	}
}

//    10000	    775683 ns/op	  243204 B/op	     180 allocs/op
func BenchmarkWS_WSClient_MemAlloc(b *testing.B) {
	defer startTestServer().Shutdown(context.TODO())
	uri := "ws://localhost:9000/ws"

	for i := 0; i < b.N; i++ {
		cli, err := NewWSClient(uri)
		if err != nil {
			b.Error(err)
		}
		cli.Close()
	}
}

//     3000	   9632127 ns/op	  242368 B/op	     176 allocs/op
func BenchmarkWS_WSClient_ConnectClose(b *testing.B) {
	defer startTestServer().Shutdown(context.TODO())
	uri := "ws://localhost:9000/ws"
	cli, err := NewWSClient(uri)
	for i := 0; i < b.N; i++ {
		if err != nil {
			b.Error(err)
		}
		cli.Close()
		err = cli.Connect()
	}
}

