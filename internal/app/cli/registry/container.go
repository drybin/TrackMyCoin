package registry

import (
	"context"
	"log"

	"github.com/drybin/TrackMyCoin/internal/adapter/webapi"
	"github.com/drybin/TrackMyCoin/internal/app/cli/config"
	"github.com/drybin/TrackMyCoin/internal/app/cli/usecase"
	"github.com/drybin/TrackMyCoin/pkg/logger"
	"github.com/drybin/TrackMyCoin/pkg/wrap"
	"github.com/go-resty/resty/v2"
)

type Container struct {
	Logger   logger.ILogger
	Usecases *Usecases
	Clean    func()
}

type Usecases struct {
	HelloWorld *usecase.HelloWorld
	Process    *usecase.Process
}

func NewContainer(
	config *config.Config,
) (*Container, error) {
	appLogger := logger.NewLogger()

	ctx := context.Background()

	// Initialize HTTP client
	httpClient := resty.New()

	// Initialize Google Sheets client
	var googleSheets *webapi.GoogleSheets
	var err error

	// Приоритет: Service Account файл > API Key
	if config.GoogleServiceAccountFile != "" {
		googleSheets, err = webapi.NewGoogleSheetsWithServiceAccount(ctx, config.GoogleServiceAccountFile)
		if err != nil {
			log.Printf("Warning: failed to create Google Sheets client with service account: %v", err)
			log.Printf("Make sure the service account file exists at: %s", config.GoogleServiceAccountFile)
		}
	} else if config.GoogleAPIKey != "" {
		googleSheets, err = webapi.NewGoogleSheetsWithAPIKey(ctx, config.GoogleAPIKey)
		if err != nil {
			return nil, wrap.Errorf("failed to create Google Sheets client with API key: %w", err)
		}
	}

	// Initialize CoinGecko client
	coinGecko := webapi.NewCoinGecko(httpClient)

	container := Container{
		Logger: appLogger,
		Usecases: &Usecases{
			HelloWorld: usecase.NewHelloWorldUsecase(),
			Process:    usecase.NewProcessUsecase(googleSheets, coinGecko, config),
		},
		Clean: func() {
		},
	}

	return &container, nil
}
