package monitoring

import "os"

func IsMonitoringEnabled() bool {
	return os.Getenv("DISABLE_MONITORING") != "true"
}
