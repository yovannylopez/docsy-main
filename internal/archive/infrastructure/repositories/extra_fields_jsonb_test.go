package repositories

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/entities"
)

func TestExtraFieldsJSONB_RoundTrip(t *testing.T) {
	src := extraFieldsJSONB{
		{Key: "contract_number", Label: "Contrato", Value: "1286243"},
		{Key: "electronic_invoice", Label: "Factura electrónica", Value: "EFIR18208812"},
	}
	v, err := src.Value()
	require.NoError(t, err)
	b, ok := v.([]byte)
	require.True(t, ok)

	var dst extraFieldsJSONB
	require.NoError(t, dst.Scan(b))
	require.Len(t, dst, 2)
	assert.Equal(t, entities.ExtraField(src[0]), entities.ExtraField(dst[0]))
}

func TestExtraFieldsJSONB_NilEmpty(t *testing.T) {
	var empty extraFieldsJSONB
	v, err := empty.Value()
	require.NoError(t, err)
	assert.Nil(t, v)

	var dst extraFieldsJSONB
	require.NoError(t, dst.Scan(nil))
	assert.Nil(t, dst)
}
