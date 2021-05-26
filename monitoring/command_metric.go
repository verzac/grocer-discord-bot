package monitoring

import (
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/verzac/grocer-discord-bot/handlers"
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
	mh        *handlers.MessageHandlerContext
}

func NewCommandMetric(cw *cloudwatch.CloudWatch, mh *handlers.MessageHandlerContext) *CommandMetric {
	return &CommandMetric{cw: cw, startTime: time.Now(), mh: mh}
}

func (cm *CommandMetric) Done() {
	command := cm.mh.GetCommand()
	if command == "" {
		return
	}
	completedIn := float64(time.Now().Sub(cm.startTime).Milliseconds())
	if CloudWatchEnabled() {
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
			log.Println(cm.mh.FmtErrMsg(err))
		}
	} else if IsMonitoringEnabled() {
		log.Printf("%s: %fms", command, completedIn)
	}
}
