package helpers

import (
	"fmt"
	"log"
	"strings"
)

// silentLogger implements fasthttp.Logger
type SilentLogger struct{}

func (s *SilentLogger) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	// drop the common "use of closed network connection" noise
	if strings.Contains(msg, "use of closed network connection") {
		return
	}
	// otherwise print as usual
	log.Printf(format, args...)
}
