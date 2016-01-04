package awsutils

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// workerq is smply a collection of a job ID and value.
type workerq struct {
	name  string
	value []string
}

/*
GetZones kicks a worker off per slice item in regions, compiles the results
from the workers into a map of region:zone and returns the map.

for example (pretty)

{
  "ap-northeast-1": [
    "ap-northeast-1a",
    "ap-northeast-1b",
    "ap-northeast-1c"
  ],
  "ap-southeast-1": [
    "ap-southeast-1a",
    "ap-southeast-1b"
  ]
}

*/
func GetZones(regions []*ec2.Region) map[string][]string {
	// build the output queue.
	q := make(chan *workerq, len(regions))

	// kick off the workers and dispatch work to each.
	for _, r := range regions {
		go func(r *ec2.Region, q chan *workerq) {
			azs := []string{}
			// ecc := ec2.New(&aws.Config{Region: aws.String(r)})
			ecc := ec2.New(session.New(), &aws.Config{Region: r.RegionName})
			resp, err := ecc.DescribeAvailabilityZones(nil)

			if err != nil {
				fmt.Println(err)
			}

			// pump the zones into the slice.
			for _, z := range resp.AvailabilityZones {
				if *z.State != "available" {
					continue
				}
				azs = append(azs, *z.ZoneName)
			}

			q <- &workerq{*r.RegionName, azs}

		}(r, q)

	}

	// read the responses from the workers.
	var (
		result *workerq
		m      = make(map[string][]string)
	)
	for i := 0; i < len(regions); i++ {
		result = <-q
		m[result.name] = result.value
	}

	return m
}
