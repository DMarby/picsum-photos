package health_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/DMarby/picsum-photos/health"
	"github.com/DMarby/picsum-photos/logger"
	"go.uber.org/zap"

	fileDatabase "github.com/DMarby/picsum-photos/database/file"
	mockDatabase "github.com/DMarby/picsum-photos/database/mock"

	mockProcessor "github.com/DMarby/picsum-photos/image/mock"
	vipsProcessor "github.com/DMarby/picsum-photos/image/vips"

	fileStorage "github.com/DMarby/picsum-photos/storage/file"
	mockStorage "github.com/DMarby/picsum-photos/storage/mock"
)

func TestHealth(t *testing.T) {
	log := logger.New(zap.ErrorLevel)
	defer log.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	imageProcessor, _ := vipsProcessor.GetInstance(ctx, log)
	storage, _ := fileStorage.New("../test/fixtures/file")
	db, _ := fileDatabase.New("../test/fixtures/file/metadata.json")

	checker := health.New(ctx, imageProcessor, storage, db)
	mockStorageChecker := health.New(ctx, imageProcessor, &mockStorage.Provider{}, db)
	mockProcessorChecker := health.New(ctx, &mockProcessor.Processor{}, storage, db)
	mockDatabaseChecker := health.New(ctx, imageProcessor, storage, &mockDatabase.Provider{})

	tests := []struct {
		Name           string
		ExpectedStatus health.Status
		Checker        *health.Checker
	}{
		{
			Name: "runs checks and returns correct status",
			ExpectedStatus: health.Status{
				Healthy:   true,
				Database:  "healthy",
				Processor: "healthy",
				Storage:   "healthy",
			},
			Checker: checker,
		},
		{
			Name: "runs checks and returns correct status with broken storage",
			ExpectedStatus: health.Status{
				Healthy:   false,
				Database:  "healthy",
				Processor: "unknown",
				Storage:   "unhealthy",
			},
			Checker: mockStorageChecker,
		},
		{
			Name: "runs checks and returns correct status with broken processor",
			ExpectedStatus: health.Status{
				Healthy:   false,
				Database:  "healthy",
				Processor: "unhealthy",
				Storage:   "healthy",
			},
			Checker: mockProcessorChecker,
		},
		{
			Name: "runs checks and returns correct status with broken database",
			ExpectedStatus: health.Status{
				Healthy:   false,
				Database:  "unhealthy",
				Processor: "unknown",
				Storage:   "unknown",
			},
			Checker: mockDatabaseChecker,
		},
	}

	for _, test := range tests {
		test.Checker.Run()
		status := test.Checker.Status()

		if !reflect.DeepEqual(status, test.ExpectedStatus) {
			t.Errorf("%s: wrong status %+v", test.Name, status)
		}
	}

	t.Run("checker runs and returns correct status", func(t *testing.T) {

	})
}
