package api

import (
	"cmp"
	"fmt"
	"io"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/gin-gonic/gin"

	"github.com/ladev707/3dbuilder-backend/internal/config"
)

func NewEngine(cfg config.Config) *gin.Engine {
	return (&Application{Config: cfg}).SetupRoutes()
}

func WriteRoutes(w io.Writer, engine *gin.Engine) error {
	routes := engine.Routes()

	slices.SortFunc(routes, func(a, b gin.RouteInfo) int {
		return cmp.Or(cmp.Compare(a.Path, b.Path), cmp.Compare(a.Method, b.Method))
	})

	writer := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(writer, "METHOD\tPATH\tHANDLER")
	for _, route := range routes {
		fmt.Fprintf(writer, "%s\t%s\t%s\n", route.Method, route.Path, handlerName(route.Handler))
	}
	if err := writer.Flush(); err != nil {
		return err
	}

	_, err := fmt.Fprintf(w, "\n%d routes\n", len(routes))
	return err
}

// Learning note: Gin reports names like
// "github.com/ladev707/3dbuilder-backend/internal/src/auth/handlers.(*Handler).Me-fm".
// Go adds "-fm" when a method is passed as a value (h.Me instead of h.Me()).
// This keeps only the part after the last "/", giving "handlers.(*Handler).Me".
func handlerName(name string) string {
	name = strings.TrimSuffix(name, "-fm")
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return name
}
