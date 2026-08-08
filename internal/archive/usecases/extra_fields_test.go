package usecases

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	domainerrors "github.com/yovannylopez/docsy-main/internal/archive/domain/errors"
)

func TestNormalizeExtraFields(t *testing.T) {
	got, err := normalizeExtraFields([]dtos.ExtraFieldDTO{
		{Key: extraKeyContract, Label: "Contrato", Value: "1286243"},
		{Key: extraKeyContract, Label: "Dup", Value: "999"},
		{Key: "", Label: "x", Value: "y"},
		{Key: extraKeyElectronicInv, Label: " Factura ", Value: " EFIR1 "},
	})
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, extraKeyContract, got[0].Key)
	assert.Equal(t, "1286243", got[0].Value)
	assert.Equal(t, extraKeyElectronicInv, got[1].Key)
	assert.Equal(t, "EFIR1", got[1].Value)
}

func TestNormalizeExtraFields_TooMany(t *testing.T) {
	in := make([]dtos.ExtraFieldDTO, maxExtraFields+1)
	for i := range in {
		in[i] = dtos.ExtraFieldDTO{
			Key:   fmt.Sprintf("key_%d", i),
			Label: "L",
			Value: "V",
		}
	}
	_, err := normalizeExtraFields(in)
	assert.ErrorIs(t, err, domainerrors.ErrTooManyExtraFields)
}
