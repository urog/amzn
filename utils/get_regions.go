package awsutils

import (
  "fmt"
  "github.com/aws/aws-sdk-go/service/ec2"
)

// GetRegions returns a slice of all AWS regions.
func GetRegions(ecc *ec2.EC2) ([]*ec2.Region, error) {
	regionlist := []*ec2.Region{}
	resp, err := ecc.DescribeRegions(nil)

	if err != nil {
		fmt.Println(err.Error())
	}

	for _, r := range resp.Regions {
		regionlist = append(regionlist, r)
	}

	return regionlist, nil
}
