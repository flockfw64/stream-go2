package stream

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	testCases := []struct {
		key             string
		secret          string
		shouldError     bool
		opts            []ClientOption
		expectedRegion  string
		expectedVersion string
	}{
		{
			shouldError: true,
		},
		{
			key: "k", secret: "s",
			expectedRegion:  "",
			expectedVersion: "",
		},
		{
			key: "k", secret: "s",
			opts:            []ClientOption{WithAPIRegion("test")},
			expectedRegion:  "test",
			expectedVersion: "",
		},
		{
			key: "k", secret: "s",
			opts:            []ClientOption{WithAPIVersion("test")},
			expectedRegion:  "",
			expectedVersion: "test",
		},
		{
			key: "k", secret: "s",
			opts:            []ClientOption{WithAPIRegion("test"), WithAPIVersion("more")},
			expectedRegion:  "test",
			expectedVersion: "more",
		},
	}
	for _, tc := range testCases {
		c, err := NewClient(tc.key, tc.secret, tc.opts...)
		if tc.shouldError {
			assert.Error(t, err)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, tc.expectedRegion, c.url.region)
		assert.Equal(t, tc.expectedVersion, c.url.version)
	}
}

func Test_makeEndpoint(t *testing.T) {
	prev := os.Getenv("STREAM_URL")
	defer os.Setenv("STREAM_URL", prev)

	testCases := []struct {
		url      *apiURL
		format   string
		env      string
		args     []interface{}
		expected string
	}{
		{
			url:      &apiURL{},
			format:   "test-%d-%s",
			args:     []interface{}{42, "asd"},
			expected: "https://api.stream-io-api.com/api/v1.0/test-42-asd?api_key=test",
		},
		{
			url:      &apiURL{},
			env:      "http://localhost:8000/api/v1.0/",
			format:   "test-%d-%s",
			args:     []interface{}{42, "asd"},
			expected: "http://localhost:8000/api/v1.0/test-42-asd?api_key=test",
		},
	}

	for _, tc := range testCases {
		os.Setenv("STREAM_URL", tc.env)
		c := &Client{url: tc.url, key: "test"}
		assert.Equal(t, tc.expected, c.makeEndpoint(tc.format, tc.args...).String())
	}
}

func TestNewClientFromEnv(t *testing.T) {
	defer func() {
		os.Setenv("STREAM_API_KEY", "")
		os.Setenv("STREAM_API_SECRET", "")
		os.Setenv("STREAM_API_REGION", "")
		os.Setenv("STREAM_API_VERSION", "")
	}()

	_, err := NewClientFromEnv()
	require.Error(t, err)

	os.Setenv("STREAM_API_KEY", "foo")
	os.Setenv("STREAM_API_SECRET", "bar")

	client, err := NewClientFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "foo", client.key)
	assert.Equal(t, "bar", client.authenticator.secret)

	os.Setenv("STREAM_API_REGION", "baz")
	client, err = NewClientFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "baz", client.url.region)

	os.Setenv("STREAM_API_VERSION", "qux")
	client, err = NewClientFromEnv()
	require.NoError(t, err)
	assert.Equal(t, "qux", client.url.version)
}

type badReader struct{}

func (badReader) Read([]byte) (int, error) { return 0, fmt.Errorf("boom") }

func Test_makeStreamError(t *testing.T) {
	testCases := []struct {
		body     io.Reader
		expected error
		apiErr   APIError
	}{
		{
			body:     nil,
			expected: fmt.Errorf("invalid body"),
		},
		{
			body:     badReader{},
			expected: fmt.Errorf("boom"),
		},
		{
			body:     strings.NewReader(`{{`),
			expected: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
		{
			body:     strings.NewReader(`{"code":"A"}`),
			expected: fmt.Errorf("json: cannot unmarshal string into Go struct field APIError.code of type int"),
		},
		{
			body:     strings.NewReader(`{"code":123, "detail":"test", "duration": "1m2s", "exception": "boom", "status_code": 456, "exception_fields": {"foo":["bar"]}}`),
			expected: fmt.Errorf("test"),
			apiErr: APIError{
				Code:       123,
				Detail:     "test",
				Duration:   Duration{time.Minute + time.Second*2},
				Exception:  "boom",
				StatusCode: 456,
				ExceptionFields: map[string][]interface{}{
					"foo": []interface{}{"bar"},
				},
			},
		},
	}
	for _, tc := range testCases {
		err := (&Client{}).makeStreamError(tc.body)
		assert.Equal(t, tc.expected.Error(), err.Error())
		if tc.apiErr.Code != 0 {
			assert.Equal(t, tc.apiErr, err)
		}
	}
}
