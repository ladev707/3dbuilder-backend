package api

import (
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/ladev707/3dbuilder-backend/internal/config"
)

func TestWriteRoutesSortsByPathThenMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	noop := func(*gin.Context) {}
	engine.POST("/b", noop)
	engine.GET("/b", noop)
	engine.GET("/a", noop)

	// strings.Builder is an io.Writer, so it captures the output in memory
	// instead of printing it to the terminal.
	var out strings.Builder
	if err := WriteRoutes(&out, engine); err != nil {
		t.Fatalf("WriteRoutes: %v", err)
	}

	// lines[0] is the header, lines[1:4] are the three routes, and the last
	// line is the summary. strings.Fields splits on spaces and ignores the
	// padding added by tabwriter.
	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	var got []string
	for _, line := range lines[1:4] {
		fields := strings.Fields(line)
		got = append(got, fields[0]+" "+fields[1])
	}
	want := []string{"GET /a", "GET /b", "POST /b"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("routes = %v, want %v", got, want)
	}
	if last := lines[len(lines)-1]; last != "3 routes" {
		t.Fatalf("summary = %q, want %q", last, "3 routes")
	}
}

// Learning note: this test builds the real router with an empty config and no
// database. If a later change made route setup touch the database, the test
// would fail (nil pointer) and catch it.
func TestNewEngineRegistersRoutesWithoutDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	routes := NewEngine(config.Config{}).Routes()

	// Learning note: a map[string]bool works as a set, so each lookup is
	// fast. Missing keys return false.
	registered := make(map[string]bool, len(routes))
	for _, route := range routes {
		registered[route.Method+" "+route.Path] = true
	}
	for _, route := range []string{"GET /health", "POST /api/v1/auth/user/login", "GET /api/v1/users"} {
		if !registered[route] {
			t.Errorf("route %s is not registered", route)
		}
	}
}
