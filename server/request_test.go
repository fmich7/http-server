package server

import (
	"net"
	"testing"
)

// mockConn sets server and client for testing
func mockConn(input string) (net.Conn, net.Conn) {
	server, client := net.Pipe()
	go func() {
		defer client.Close()
		client.Write([]byte(input)) // Send data to the server end
	}()
	return server, client
}

func TestReadRequest(t *testing.T) {
	t.Run("normal data", func(t *testing.T) {
		input := "ale lekki test!"
		server, _ := mockConn(input)
		defer server.Close()

		data, err := ReadRequest(server)
		if err != nil {
			t.Fatal("Expected no error")
		}

		if string(data) != input {
			t.Errorf("Expected [%s], got [%s]\n", input, string(data))
		}
	})

	t.Run("random", func(t *testing.T) {
		input := "t\r\n\r\na"
		server, client := mockConn(input)
		defer client.Close()
		defer server.Close()

		go func() {
			client.Close()
		}()

		_, err := ReadRequest(server)
		if err != nil {
			t.Fatalf("Expected EOF error or nil, got: %v", err)
		}
	})
	t.Run("empty", func(t *testing.T) {
		input := ""
		server, client := mockConn(input)
		defer client.Close()
		defer server.Close()

		_, err := ReadRequest(server)
		if err != nil {
			t.Fatalf("Expected EOF error or nil, got: %v", err)
		}
	})
}

func TestParseRequest(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		input := ""
		server, _ := mockConn(input)
		defer server.Close()

		httpRequest, err := ParseRequest(server)
		t.Log(httpRequest)
		if err == nil {
			t.Fatal("Expected an error", err)
		}
	})

	t.Run("normal request", func(t *testing.T) {
		input := "GET /index.html HTTP/1.1\r\n\r\n"
		server, _ := mockConn(input)
		defer server.Close()

		got, err := ParseRequest(server)
		if err != nil {
			t.Fatal(err)
		}

		want := HTTPRequest{
			"GET",
			"/index.html",
			"HTTP/1.1",
			make(map[string]string),
			make([]byte, 0),
		}

		if !isEqualHTTPRequest(want, got) {
			t.Fatalf("Want %v, want %v ", want, got)
		}
	})

	t.Run("normal request", func(t *testing.T) {
		input := "GET /index.html HTTP/1.1\r\nHost: example.com\r\n\r\nHello, Go!"
		server, _ := mockConn(input)
		defer server.Close()

		got, err := ParseRequest(server)
		if err != nil {
			t.Fatal(err)
		}

		want := HTTPRequest{
			Method:          "GET",
			URL:             "/index.html",
			ProtocolVersion: "HTTP/1.1",
			Headers:         map[string]string{"Host": "exampleeeeee.com"},
			Body:            []byte("Hello, Go!"),
		}

		if isEqualHTTPRequest(want, got) {
			t.Fatal("These requests does not contain same values!")
		}
	})

}

func TestHTTPRequestComparison(t *testing.T) {
	tests := []struct {
		name  string
		want  HTTPRequest
		got   HTTPRequest
		equal bool
	}{
		{
			name: "Equal requests",
			want: HTTPRequest{
				Method:          "GET",
				URL:             "/index.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "example.com"},
				Body:            []byte{},
			},
			got: HTTPRequest{
				Method:          "GET",
				URL:             "/index.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "example.com"},
				Body:            []byte{},
			},
			equal: true,
		},
		{
			name: "Different methods",
			want: HTTPRequest{
				Method:          "GET",
				URL:             "/index.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "example.com"},
				Body:            []byte{},
			},
			got: HTTPRequest{
				Method:          "POST",
				URL:             "/index.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "example.com"},
				Body:            []byte{},
			},
			equal: false,
		},
		{
			name: "Different headers",
			want: HTTPRequest{
				Method:          "GET",
				URL:             "/index.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "example.com"},
				Body:            []byte{},
			},
			got: HTTPRequest{
				Method:          "GET",
				URL:             "/index.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "different.com"},
				Body:            []byte{},
			},
			equal: false,
		},
		{
			name: "Different body",
			want: HTTPRequest{
				Method:          "GET",
				URL:             "/index.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "example.com"},
				Body:            []byte("Hello, world!"),
			},
			got: HTTPRequest{
				Method:          "GET",
				URL:             "/index.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "example.com"},
				Body:            []byte("Hello, Go!"),
			},
			equal: false,
		},
		{
			name: "Different URL",
			want: HTTPRequest{
				Method:          "GET",
				URL:             "/index.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "example.com"},
				Body:            []byte{},
			},
			got: HTTPRequest{
				Method:          "GET",
				URL:             "/about.html",
				ProtocolVersion: "HTTP/1.1",
				Headers:         map[string]string{"Host": "example.com"},
				Body:            []byte{},
			},
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := isEqualHTTPRequest(tt.want, tt.got)
			if gotErr != tt.equal {
				t.Errorf("isEqualHTTPRequest() = %v, wantErr %v", gotErr, tt.equal)
			}
		})
	}
}