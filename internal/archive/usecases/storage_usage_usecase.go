package usecases

import (
	"context"
	"fmt"
	"strings"

	"github.com/yovannylopez/docsy-main/internal/archive/domain/dtos"
	"github.com/yovannylopez/docsy-main/internal/archive/domain/ports"
)

const storagePercentCap = 100

// GetStorageUsageUseCase aggregates attachment bytes for a user's workspaces.
type GetStorageUsageUseCase struct {
	fileRepo   ports.DocumentFileRepository
	quotaBytes int64
}

// NewGetStorageUsageUseCase creates the use case.
func NewGetStorageUsageUseCase(fileRepo ports.DocumentFileRepository, quotaBytes int64) *GetStorageUsageUseCase {
	return &GetStorageUsageUseCase{fileRepo: fileRepo, quotaBytes: quotaBytes}
}

// Execute returns used bytes and soft quota for the sidebar indicator.
func (uc *GetStorageUsageUseCase) Execute(ctx context.Context, userID string) (*dtos.StorageUsageResponse, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}

	used, err := uc.fileRepo.SumSizeBytesForUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("sum storage usage: %w", err)
	}

	quota := uc.quotaBytes
	if quota < 0 {
		quota = 0
	}

	percent := 0
	if quota > 0 {
		percent = int((used * int64(storagePercentCap)) / quota)
		if percent > storagePercentCap {
			percent = storagePercentCap
		}
	}

	return &dtos.StorageUsageResponse{
		UsedBytes:  used,
		QuotaBytes: quota,
		Percent:    percent,
	}, nil
}
