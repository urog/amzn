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
	instances = make(map[string]map[string]string)
	out       []byte
)

// getInstances grabs a list of instances that have a "Name" tag, and optionally
// accepts a regular expression to filter the results
// TODO: add option to dump ALL instances regardless of tags.
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
			instance := make(map[string]string)
			// ignore nodes without an IP address
			if i.PrivateIpAddress == nil {
				continue
			}
			for _, t := range i.Tags {
				var nametag string
				// only take nodes with a "Name" tag
				if *t.Key != "Name" {
					continue
				} else {
					match, _ := regexp.MatchString(filter, *t.Value)
					if match {
						nametag = *t.Value
					} else {
						continue
					}
				}
				// pump some metadata into our instance map
				instance["name"] = nametag
				instance["local-ip"] = *i.PrivateIpAddress
				if i.PublicIpAddress != nil {
					instance["public-ip"] = *i.PublicIpAddress
				}
				instance["zone"] = *i.Placement.AvailabilityZone
				// add the instance map into the instances map
				instances[*i.InstanceId] = instance
			}
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
