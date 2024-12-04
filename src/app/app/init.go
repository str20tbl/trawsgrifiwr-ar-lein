package app

import (
	"app/app/appJobs"
	"fmt"
	_ "github.com/revel/modules"
	"github.com/revel/revel"
	"time"
)

var (
	// AppVersion revel app version (ldflags)
	AppVersion string

	// BuildTime revel app build-time (ldflags)
	BuildTime string
)

func init() {
	// Filters is the default set of global filters.
	revel.Filters = []revel.Filter{
		revel.PanicFilter,             // Recover from panics and display an error page instead.
		revel.RouterFilter,            // Use the routing table to select the right Action
		revel.FilterConfiguringFilter, // A hook for adding or removing per-Action filters.
		revel.ParamsFilter,            // Parse parameters into Controller.Params.
		revel.SessionFilter,           // Restore and write the session cookie.
		revel.FlashFilter,             // Restore and write the flash cookie.
		revel.ValidationFilter,        // Restore kept validation errors and save new ones from cookie.
		revel.I18nFilter,              // Resolve the requested language
		HeaderFilter,                  // Add some security based headers
		revel.InterceptorFilter,       // Run interceptors around the action.
		revel.CompressFilter,          // Compress the result.
		revel.BeforeAfterFilter,       // Call the before and after filter functions
		revel.ActionInvoker,           // Invoke the action.
	}
	revel.TemplateFuncs["percent"] = func(input, total int) string {
		sum := 0.0
		if total > 0 {
			sum = float64(input) / float64(total) * 100
		}
		return fmt.Sprintf("%.0f", sum)
	}

	revel.TemplateFuncs["len"] = func(input appJobs.Transcript) int {
		return len(input.Segments) - 1
	}

	revel.TemplateFuncs["notIn"] = func(uuid string, blackList []string) (fnd bool) {
		for _, item := range blackList {
			if uuid == item {
				fnd = true
				break
			}
		}
		return !fnd
	}

	revel.TemplateFuncs["prevTime"] = func(input appJobs.Transcript, index int) float64 {
		if index == 0 {
			return 0.0
		}
		return input.Segments[index-1].Start
	}

	revel.TemplateFuncs["nextTime"] = func(input appJobs.Transcript, index int) float64 {
		if index+1 >= len(input.Segments)-1 {
			return input.Segments[len(input.Segments)-1].Start
		}
		return input.Segments[index+1].Start
	}

	revel.TemplateFuncs["add"] = func(a, b int) int {
		return a + b
	}

	revel.TemplateFuncs["div"] = func(input appJobs.Transcript, b int) int {
		return int(float64(len(input.Segments)) / float64(b))
	}

	revel.TemplateFuncs["s2m"] = func(inSeconds float64) string {
		return time.Duration(inSeconds * float64(time.Second)).String()
	}

	revel.TemplateFuncs["addTime"] = func(a, b float64) float64 {
		return a + b
	}

	// Register startup functions with OnAppStart
	// revel.DevMode and revel.RunMode only work inside of OnAppStart. See Example Startup Script
	// ( order dependent )
	// revel.OnAppStart(ExampleStartupScript)
	// revel.OnAppStart(InitDB)
	// revel.OnAppStart(FillCache)
}

// HeaderFilter adds common security headers
// There is a full implementation of a CSRF filter in
// https://github.com/revel/modules/tree/master/csrf
var HeaderFilter = func(c *revel.Controller, fc []revel.Filter) {
	c.Response.Out.Header().Add("X-Frame-Options", "SAMEORIGIN")
	c.Response.Out.Header().Add("X-XSS-Protection", "1; mode=block")
	c.Response.Out.Header().Add("X-Content-Type-Options", "nosniff")
	c.Response.Out.Header().Add("Referrer-Policy", "strict-origin-when-cross-origin")

	fc[0](c, fc[1:]) // Execute the next filter stage.
}

//func ExampleStartupScript() {
//	// revel.DevMod and revel.RunMode work here
//	// Use this script to check for dev mode and set dev/prod startup scripts here!
//	if revel.DevMode == true {
//		// Dev mode
//	}
//}
