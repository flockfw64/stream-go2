package stream_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	stream "github.com/flockfw64/stream-go2/v7"
)

func TestDurationMarshalUnmarshalJSON(t *testing.T) {
	dur := stream.Duration{Duration: 33 * time.Second}
	data := []byte(`"33s"`)
	marshaled, err := json.Marshal(dur)
	assert.NoError(t, err)
	assert.Equal(t, data, marshaled)
	var out stream.Duration
	err = json.Unmarshal(marshaled, &out)
	assert.NoError(t, err)
	assert.Equal(t, dur, out)
}

func TestTimeMarshalUnmarshalJSON(t *testing.T) {
	tt, _ := time.Parse("2006-Jan-02", "2013-Feb-03")
	st := stream.Time{Time: tt}
	data := []byte(`"2013-02-03T00:00:00"`)
	marshaled, err := json.Marshal(st)
	require.NoError(t, err)
	require.Equal(t, data, marshaled)
	var out stream.Time
	err = json.Unmarshal(marshaled, &out)
	assert.NoError(t, err)
	assert.Equal(t, st, out)
}

func TestEnrichedActivityMarshal(t *testing.T) {
	e := stream.EnrichedActivity{
		Actor: stream.Data{
			ID:    "my_id",
			Extra: map[string]interface{}{"a": 1, "b": "c"},
		},
		ReactionCounts: map[string]int{
			"comment": 1,
		},
		Score: 100.0,
	}

	b, err := json.Marshal(e)
	require.NoError(t, err)
	require.JSONEq(t, `{"actor": {"id":"my_id","data":{"a":1,"b":"c"}},"reaction_counts":{"comment":1},"score":100.0}`, string(b))
}
