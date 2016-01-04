package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/urog/amzn/utils"
)

// This program spits out a pretty-printed JSON document describing
// the exposed availability zones in all regions to your AWS account

// Dump writes results to the current working dir
func Dump(result []byte) {
	fileName := "zones.json.tf"
	fmt.Printf("Writing output to \"%v\"...\n", fileName)
	err := ioutil.WriteFile(fileName, result, 0644)
	Check(err)
}

// Check will panic on errors
func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	ecc := ec2.New(session.New(), &aws.Config{
		Region: aws.String(os.Getenv("EC2_REGION"))},
	)

	regions, err := awsutils.GetRegions(ecc)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Finding all the zones...")
	zoneList := awsutils.GetZones(regions)

	zonesAsString := make(map[string]string)
	zoneCount := make(map[string]int)
	zoneIdentifiers := make(map[string]string)
	// Set the default values for our TF zones Var
	defaultZonesAsString := make(map[string]map[string]string)
	defaultZoneCount := make(map[string]map[string]int)
	defaultZoneIdentifiers := make(map[string]map[string]string)
	// simple rex to extract the zone identifier
	re, _ := regexp.Compile(`[a-f]{1}$`)

	for r, z := range zoneList {
		zid := []string{}
		// TF can't have variables in a list/array form.
		zonesAsString[r] = strings.Join(z, ",")
		zoneCount[r] = len(z)

		for _, id := range z {
			res := re.FindAllStringSubmatch(id, -1)
			zid = append(zid, res[0][0]) // the regex returns a [][]string type
		}

		zoneIdentifiers[r] = strings.Join(zid, ",")

	}

	type (
		Zones struct {
			Zones           map[string]map[string]string `json:"zones"`
			ZoneCount       map[string]map[string]int    `json:"zone_count"`
			ZoneIdentifiers map[string]map[string]string `json:"zone_identifiers"`
		}

		TFVar struct {
			Variable Zones `json:"variable"`
		}
	)

	defaultZonesAsString["default"] = zonesAsString
	defaultZoneCount["default"] = zoneCount
	defaultZoneIdentifiers["default"] = zoneIdentifiers

	zones := &Zones{
		Zones:           defaultZonesAsString,
		ZoneCount:       defaultZoneCount,
		ZoneIdentifiers: defaultZoneIdentifiers,
	}

	tfVar := &TFVar{
		Variable: *zones,
	}

	// print out the results.
	var out []byte
	out, _ = json.MarshalIndent(&tfVar, "", "  ")
	Dump(out)
	// fmt.Println(string(out))

}
