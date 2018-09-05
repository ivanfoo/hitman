package main

import (
	"os"
	"time"
	"flag"
	"log"

	"github.com/shirou/gopsutil/cpu"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
)
var (
	isDisabled = flag.Bool("disabled", false, "Disable the service")
	minCpuUsage = flag.Float64("mincpu", 20, "Minimal CPU usage to mark the instance as busy")
	maxIdleCount =  flag.Int("maxidlecount", 30, "Max number of times the instance can be flagged as idle")
)

func main() {
	flag.Parse()

	if *isDisabled {
		log.Println("Service is disabled. Stopping...\n")
		os.Exit(0)
	}

	awsClient := session.Must(session.NewSession())
	metadataClient := ec2metadata.New(awsClient)
	autoscalingClient := autoscaling.New(awsClient)

	if !metadataClient.Available() {
		log.Println("EC2 metadata not available. Stopping....\n")
		os.Exit(0)

	}

	metadata, _ := metadataClient.GetInstanceIdentityDocument()
	idleCount := 0

	for {
		cpuUsage, _ := cpu.Percent(time.Duration(60)*time.Second, false)
		avgCpuUsage := cpuUsage[0]

		if avgCpuUsage > *minCpuUsage {
			idleCount = 0
			continue
		}

		idleCount += 1
		log.Println("Flagged instance as idle. Total iddle count: (%d)", idleCount)

		if idleCount < *maxIdleCount {
			continue
		}

		log.Println("CPU usage is %f %%: terminating...\n", avgCpuUsage)

		_, err := autoscalingClient.TerminateInstanceInAutoScalingGroup(
			&autoscaling.TerminateInstanceInAutoScalingGroupInput{
				InstanceId:                     aws.String(metadata.InstanceID),
				ShouldDecrementDesiredCapacity: aws.Bool(true),
			})

		if err != nil {
			log.Println(err)
		}
	}
}
