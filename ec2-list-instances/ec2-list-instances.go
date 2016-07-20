package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

var (
	region    = os.Getenv("EC2_REGION")
	instances []instance
	out       []byte
)

type (
	instance struct {
		Name             string `json:"name"`
		PublicIp         string `json:"public-ip"`
		PrivateIp        string `json:"private-ip"`
		AvailabilityZone string `json:"zone"`
	}
)

// getInstances grabs a list of instances in a particular region and,
// optionally, accepts a regular expression to filter the results
func getInstances(region *string, filter string) {
	// initialise a connection to the EC2 API
	svc := ec2.New(session.New(), &aws.Config{
		Region: region})
	// call the DescribeInstances Operation
	resp, err := svc.DescribeInstances(nil)
	if err != nil {
		panic(err)
	}
	for r := range resp.Reservations {
		for _, i := range resp.Reservations[r].Instances {
			inst := instance{}

			// ignore nodes without an IP address
			if i.PrivateIpAddress == nil {
				continue
			}

			hasName := false
			var nameTag string

			for _, t := range i.Tags {
				if *t.Key != "Name" {
					continue
				} else {
					hasName = true
					nameTag = *t.Value
				}

			}

			// pump some metadata into our instance map
			if hasName {
				match, _ := regexp.MatchString(filter, nameTag)
				if match {
					inst.Name = nameTag
				} else {
					continue
				}
			}

			inst.PrivateIp = *i.PrivateIpAddress

			if i.PublicIpAddress != nil {
				inst.PublicIp = *i.PublicIpAddress
			}

			inst.AvailabilityZone = *i.Placement.AvailabilityZone

			// add the instance struct into the instances array
			instances = append(instances, inst)

		}
	}
	// pretty print the result as json
	out, _ = json.MarshalIndent(&instances, "", "  ")
	fmt.Println(string(out))
}

func main() {
	filter := flag.String("filter", ".*", "An expression to filter results")
	region := flag.String("region", region, "The AWS region to connect")
	flag.Parse()
	getInstances(region, *filter)
}
