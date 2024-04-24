package template

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"smr/pkg/config"
	"smr/pkg/container"
	"smr/pkg/logger"
	"smr/pkg/runtime"
)

func RenderResource(templateDir string, resourceDir string, file string, config *config.Config, runtime runtime.Runtime) {
	/*
		if file != "" {
			combinator := &Combinator{
				config.Configuration,
				runtime,
			}

			filePath := fmt.Sprintf("%s/%s", templateDir, file)
			p, err := filepath.Abs(filePath)

			CustomFunctions := template.FuncMap{
				"Normalize": func(str string) string {
					return strings.ToLower(strings.Replace(str, ".", "_", 1))
				},
			}

			fi, err := os.OpenFile(fmt.Sprintf("%s/%s", resourceDir, file), os.O_CREATE|os.O_WRONLY, os.ModePerm)
			if err != nil {
				logger.Log.Fatal("Failed to open rendered file",
					zap.String("error", err.Error()))
			}
			defer fi.Close()

			t := template.Must(template.New(filepath.Base(p)).Funcs(CustomFunctions).ParseFiles(filePath))
			err = t.Execute(fi, combinator)
			if err != nil {
				panic(err)
			}
		}
	*/
}

func GetEnvs(container *container.Container, runtime runtime.Runtime) ([]string, error) {
	/*
		path := fmt.Sprintf("%s/%s", container.Static.ResourcesDir, ".envs")

		if _, err := os.Stat(path); err == nil {
			file, err := os.Open(path)
			if err != nil {
				return nil, err
			}
			defer file.Close()

			var lines []string
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				lines = append(lines, scanner.Text())
			}
			return lines, scanner.Err()
		}
	*/
	return nil, nil
}

func CreateOrClearRenderDir(renderDir string, image string, config *config.Config) {
	path := fmt.Sprintf("%s", renderDir)

	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			logger.Log.Fatal("Failed to create render directory",
				zap.String("error", err.Error()))
		}
	} else {
		err := os.RemoveAll(path)
		if err != nil {
			logger.Log.Fatal("Failed to clear render directory",
				zap.String("error", err.Error()))
		}

		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			logger.Log.Fatal("Failed to create render directory",
				zap.String("error", err.Error()))
		}
	}
}
