package usecases

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yovannylopez/docsy-main/pkg/constants"
)

func TestGetStorageUsageUseCase_SumsAndCapsPercent(t *testing.T) {
	fileRepo := &mockFileRepo{}
	fileRepo.On("SumSizeBytesForUser", context.Background(), "user-1").Return(int64(1000), nil)
	uc := NewGetStorageUsageUseCase(fileRepo, 1000)

	got, err := uc.Execute(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, int64(1000), got.UsedBytes)
	assert.Equal(t, int64(1000), got.QuotaBytes)
	assert.Equal(t, 100, got.Percent)
}

func TestGetStorageUsageUseCase_RequiresUserID(t *testing.T) {
	uc := NewGetStorageUsageUseCase(&mockFileRepo{}, constants.DefaultStorageQuotaBytes)
	_, err := uc.Execute(context.Background(), " ")
	require.Error(t, err)
}

func TestGetStorageUsageUseCase_OverQuotaCapsAt100(t *testing.T) {
	fileRepo := &mockFileRepo{}
	fileRepo.On("SumSizeBytesForUser", context.Background(), "user-1").Return(int64(2500), nil)
	uc := NewGetStorageUsageUseCase(fileRepo, 1000)

	got, err := uc.Execute(context.Background(), "user-1")
	require.NoError(t, err)
	assert.Equal(t, 100, got.Percent)
	assert.Equal(t, int64(2500), got.UsedBytes)
}
