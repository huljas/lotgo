package client

import (
	"testing"
	"context"
)

//   200000	     44291 ns/op	    2277 B/op	      21 allocs/op
func BenchmarkHTTP_HttpClient_Get(b *testing.B) {
	defer startTestServer().Shutdown(context.TODO())
	cli := NewHttpClient()
	uri := "http://localhost:9000/test/ping"
	for i := 0; i < b.N; i++ {
		status, _, err := cli.Get(uri)
		if err != nil {
			b.Error(err)
		}
		if status != 200 {
			b.Error("Invalid status ", status)
		}
	}
}

//   100000	     50848 ns/op	    2999 B/op	      29 allocs/op
func BenchmarkHTTP_HttpClient_GetJSON(b *testing.B) {
	defer startTestServer().Shutdown(context.TODO())
	cli := NewHttpClient()
	uri := "http://localhost:9000/test/v1/users/bob"
	resp := &User{}
	for i := 0; i < b.N; i++ {
		status, err := cli.GetJSON(uri, resp)
		if err != nil {
			b.Error(err)
		}
		if status != 200 {
			b.Error("Invalid status ", status)
		}
		if resp.Age != 34 {
			b.Error("Wrong age")
		}
	}
}

//   100000	     58922 ns/op	    3354 B/op	      35 allocs/op
func BenchmarkHTTP_HttpClient_PostJSON(b *testing.B) {
	defer startTestServer().Shutdown(context.TODO())
	cli := NewHttpClient()
	uri := "http://localhost:9000/test/v1/users/bob"
	user := &User{Name: "Bill", Age: 11, Address: "Foo bar"}
	resp := &User{}
	for i := 0; i < b.N; i++ {
		status, err := cli.PostJSON(uri, user, resp)
		if err != nil {
			b.Error(err)
		}
		if status != 200 {
			b.Error("Invalid status ", status)
		}
	}
}

//   100000	     62418 ns/op	    3349 B/op	      35 allocs/op
func BenchmarkHTTP_HttpClient_PutJSON(b *testing.B) {
	defer startTestServer().Shutdown(context.TODO())
	cli := NewHttpClient()
	uri := "http://localhost:9000/test/v1/users/bob"
	user := &User{Name: "Bill", Age: 11, Address: "Foo bar"}
	resp := &User{}
	for i := 0; i < b.N; i++ {
		status, err := cli.PutJSON(uri, user, resp)
		if err != nil {
			b.Error(err)
		}
		if status != 200 {
			b.Error("Invalid status ", status)
		}
	}
}

//   100000	     65939 ns/op	    3033 B/op	      30 allocs/op
func BenchmarkHTTP_HttpClient_DeleteJSON(b *testing.B) {
	defer startTestServer().Shutdown(context.TODO())
	cli := NewHttpClient()
	uri := "http://localhost:9000/test/v1/users/bob"
	resp := &User{}
	for i := 0; i < b.N; i++ {
		status, err := cli.DeleteJSON(uri, resp)
		if err != nil {
			b.Error(err)
		}
		if status != 200 {
			b.Error("Invalid status")
		}
	}
}
