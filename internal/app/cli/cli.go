package cli

import (
    "log"
    "os"
    
    "github.com/drybin/TrackMyCoin/internal/app/cli/config"
    "github.com/drybin/TrackMyCoin/internal/app/cli/registry"
    "github.com/drybin/TrackMyCoin/internal/presentation/command"
    "github.com/joho/godotenv"
    cliV2 "github.com/urfave/cli/v2"
)

const cliAppDesc = "cli tool for TrackMyCoin"

// example call go run --race ./cmd/cli/... hello-world
func Run(config *config.Config) error {
    if err := godotenv.Load(); err != nil {
        log.Println(err)
    }
    
    cnt, err := registry.NewContainer(config)
    if err != nil {
        log.Fatal("failed to create cli container", err)
    }
    
    app := cliV2.NewApp()
    app.Name = config.ServiceName
    app.Usage = cliAppDesc
    app.Commands = []*cliV2.Command{
        command.NewHelloWorldCommand(cnt.Usecases.HelloWorld),
        command.NewProcessCommand(cnt.Usecases.Process),
    }
    
    return app.Run(os.Args)
}
