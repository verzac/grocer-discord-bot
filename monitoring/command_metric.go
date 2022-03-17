package monitoring

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"go.uber.org/zap"
)

var (
	metricName    = "CommandResponseTime"
	namespace     = "grocer-discord-bot"
	unit          = cloudwatch.StandardUnitMilliseconds
	dimensionName = "Command"
)

func CloudWatchEnabled() bool {
	return os.Getenv("CW_ENABLED") == "true"
}

type CommandMetric struct {
	cw        *cloudwatch.CloudWatch
	startTime time.Time
	logger    *zap.Logger
	command   string
}

func NewCommandMetric(cw *cloudwatch.CloudWatch, command string, logger *zap.Logger) *CommandMetric {
	return &CommandMetric{cw: cw, startTime: time.Now(), command: command, logger: logger.Named("metric")}
}

func (cm *CommandMetric) Done() {
	command := cm.command
	if command == "" {
		return
	}
	completedIn := float64(time.Now().Sub(cm.startTime).Milliseconds())
	if CloudWatchEnabled() && cm.cw != nil {
		if _, err := cm.cw.PutMetricData(&cloudwatch.PutMetricDataInput{
			MetricData: []*cloudwatch.MetricDatum{
				{
					MetricName: &metricName,
					Value:      &completedIn,
					Unit:       &unit,
					Dimensions: []*cloudwatch.Dimension{
						{
							Name:  &dimensionName,
							Value: &command,
						},
					},
				},
			},
			Namespace: &namespace,
		}); err != nil {
			cm.logger.Error("Cannot send metrics to CloudWatch.", zap.Error(err))
		}
	} else if IsMonitoringEnabled() {
		cm.logger.Info(fmt.Sprintf("%s: %fms", command, completedIn))
	}
}
