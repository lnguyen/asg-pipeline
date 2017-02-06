package main

import (
	"fmt"
	"os"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

func main() {

	c := &cfclient.Config{
		ApiAddress:        os.Getenv("CF_API"),
		Username:          os.Getenv("CF_USERNAME"),
		Password:          os.Getenv("CF_PASSWORD"),
		SkipSslValidation: os.Getenv("CF_SKIP_SSL") != "",
	}
	client, err := cfclient.NewClient(c)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cfSecGroups, err := getCFSecGroups(client)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	expectedSecGroups, err := ReadSecGroupFolder(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, cfSecGroup := range cfSecGroups {
		if _, ok := expectedSecGroups[cfSecGroup.Name]; !ok {
			fmt.Println("Deleting " + cfSecGroup.Name)
			err := client.DeleteSecGroup(cfSecGroup.Guid)
			if err != nil {
				fmt.Println(cfSecGroup)
			}
		}
	}

	for _, expectedSecGroup := range expectedSecGroups {
		err := createOrUpdateSecGroup(client, expectedSecGroup)
		if err != nil {
			fmt.Println(err)
		}
	}

}

func createOrUpdateSecGroup(client *cfclient.Client, secGroup SecGroup) error {
	foundSecGroup, err := client.GetSecGroupByName(secGroup.Name)
	// Secgroup found just need to update
	if err == nil {
		fmt.Println("Updating " + foundSecGroup.Name)
		_, secGroupErr := client.UpdateSecGroup(foundSecGroup.Guid, foundSecGroup.Name, secGroup.Rules, nil)
		if secGroupErr != nil {
			return secGroupErr
		}
	} else {
		fmt.Println("Creating " + secGroup.Name)
		newSecGroup, secGroupErr := client.CreateSecGroup(secGroup.Name, secGroup.Rules, nil)
		if secGroupErr != nil {
			return secGroupErr
		}
		if secGroup.IsGlobal() {
			bindGroupErr := client.BindRunningSecGroup(newSecGroup.Guid)
			if bindGroupErr != nil {
				return bindGroupErr
			}
			bindGroupErr = client.BindStagingSecGroup(newSecGroup.Guid)
			if bindGroupErr != nil {
				return bindGroupErr
			}
		} else { // Set Sec group to org/space
			org, orgErr := client.GetOrgByName(secGroup.Org())
			if orgErr != nil {
				fmt.Println("Unable to find org " + secGroup.Org())
				return orgErr
			}
			query := make(map[string]string)
			query["name"] = secGroup.Space()
			query["organization_guid"] = org.Guid
			spaces, spaceErr := client.ListSpacesByQuery(query)
			if spaceErr != nil {
				fmt.Println("Unable to find space " + secGroup.Space())
				return spaceErr
			}
			if len(spaces) != 1 {
				return fmt.Errorf("Found more spaces then needed")

			}
			bindGroupErr := client.BindSecGroup(newSecGroup.Guid, spaces[0].Guid)
			if bindGroupErr != nil {
				return bindGroupErr
			}
		}
	}
	return nil
}

func getCFSecGroups(client *cfclient.Client) ([]cfclient.SecGroup, error) {
	var secGroups []cfclient.SecGroup
	secGroups, err := client.ListSecGroups()
	if err != nil {
		return nil, err
	}
	for _, secGroup := range secGroups {
		secGroups = append(secGroups, secGroup)
	}
	return secGroups, nil
}
