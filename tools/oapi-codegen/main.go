//go:generate mkdir -p ../../api/openapi
//go:generate go run ./main.go

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"unicode"
	"unsafe"

	"github.com/oapi-codegen/oapi-codegen/v2/pkg/codegen"
	"github.com/oapi-codegen/oapi-codegen/v2/pkg/util"
)

var (
	openapiPath = "../../openapi.yaml"
	outputPath  = "../../api/openapi/openapi_gen.go"
	config      = codegen.Configuration{
		PackageName: "openapi",
		Generate: codegen.GenerateOptions{
			GinServer:    true,
			Strict:       true,
			Models:       true,
			EmbeddedSpec: true,
		},
		OutputOptions: codegen.OutputOptions{
			UserTemplates: map[string]string{
				"strict/strict-interface.tmpl": "./strict-interface.tmpl",
			},
			SkipPrune: true,
		},
	}
)

func errExit(format string, args ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format = format + "\n"
	}
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}

func main() {
	config = config.UpdateDefaults()
	if err := config.Validate(); err != nil {
		errExit("configuration error: %v\n", err)
	}

	swagger, err := util.LoadSwaggerWithOverlay(openapiPath, util.LoadSwaggerWithOverlayOpts{Strict: true})
	if err != nil {
		errExit("error loading swagger spec in %s\n: %s\n", openapiPath, err)
	}

	code, err := codegen.Generate(swagger, config)
	if err != nil {
		errExit("error generating code: %s\n", err)
	}

	if outputPath != "" {
		err = os.WriteFile(outputPath, []byte(code), 0o644)
		if err != nil {
			errExit("error writing generated code to file: %s\n", err)
		}
	} else {
		fmt.Print(code)
	}
}

func init() {
	codegen.TemplateFunctions["correctSetCookieName"] = func(s string) string {
		data := []byte(s)
		if strings.HasPrefix(s, "SetCookie") {
			data[9] = byte(unicode.ToUpper(rune(data[9])))
		} else if strings.HasPrefix(s, "UnsetCookie") {
			data[11] = byte(unicode.ToUpper(rune(data[11])))
		}
		return unsafe.String(unsafe.SliceData(data), len(data))
	}
	codegen.TemplateFunctions["parseSetCookie"] = func(s string) *http.Cookie {
		parts := strings.SplitN(s, "|", 3)
		c, err := http.ParseSetCookie("x=x;" + parts[2])
		if err != nil {
			panic(fmt.Errorf("failed to parse set cookie: %w", err))
		}
		c.Name = parts[1]
		return c
	}
	codegen.TemplateFunctions["hasPrefix"] = func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	}
}
