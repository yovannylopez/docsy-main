package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJSONB_Value_NilReceiver(t *testing.T) {
	var j JSONB
	v, err := j.Value()
	require.NoError(t, err)
	require.Nil(t, v)
}

func TestJSONB_Value_EmptyMap(t *testing.T) {
	j := JSONB{}
	v, err := j.Value()
	require.NoError(t, err)
	require.NotNil(t, v)
}

func TestJSONB_Value_NonEmpty(t *testing.T) {
	j := JSONB{"k": "v"}
	v, err := j.Value()
	require.NoError(t, err)
	require.NotNil(t, v)
}

func TestJSONB_Scan_Nil(t *testing.T) {
	var j JSONB
	require.NoError(t, j.Scan(nil))
	require.Nil(t, j)
}

func TestJSONB_Scan_Bytes(t *testing.T) {
	var j JSONB
	require.NoError(t, j.Scan([]byte(`{"a":1}`)))
	require.Equal(t, float64(1), j["a"])
}

func TestJSONB_Scan_String(t *testing.T) {
	var j JSONB
	require.NoError(t, j.Scan(`{"b":"c"}`))
	require.Equal(t, "c", j["b"])
}
