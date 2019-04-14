package imageutils

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/pkg/archive"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

var logger, _ = zap.NewDevelopment()
var slogger = logger.Sugar()

func createImage(imageRegistry, imageName, dockerfilePath string) {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		slogger.Fatal(err, " :unable to init client")
	}

	contextDir, dockerFile := filepath.Split(dockerfilePath)

	reader, err := archive.TarWithOptions(contextDir, &archive.TarOptions{})

	imageBuildResponse, err := cli.ImageBuild(
		ctx,
		reader,
		types.ImageBuildOptions{
			Context:    reader,
			Tags:       []string{filepath.Join(imageRegistry, imageName)},
			Dockerfile: dockerFile,
			Remove:     true})
	if err != nil {
		slogger.Fatal(err, " :unable to build docker image")
	}
	defer imageBuildResponse.Body.Close()
	_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
	if err != nil {
		slogger.Fatal(err, " :unable to read image build response")
	}
}
