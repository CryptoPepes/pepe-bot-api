package util

import (
	"io/ioutil"
	"fmt"
	"os"
	"encoding/json"
	"github.com/schollz/closestmatch"
	"strings"
)

var name2id map[string]string
var id2name map[string]string
var categorizedNameList map[string][]string
var categorizedIdList map[string][]string

func init() {
	name2id, id2name = LoadNameMapping()

	categorizedNameList = make(map[string][]string)
	categorizedIdList = make(map[string][]string)

	for id, value := range id2name {
		idParts := strings.Split(id, ">")
		category := idParts[0]

		catNameList := categorizedNameList[category]
		if catNameList == nil {
			catNameList = make([]string, 0)
		}
		catNameList = append(catNameList, value)
		// Appending may have changed the slice memory location, reset map value
		categorizedNameList[category] = catNameList

		catIdList := categorizedIdList[category]
		if catIdList == nil {
			catIdList = make([]string, 0)
		}
		catIdList = append(catIdList, id)
		// Appending may have changed the slice memory location, reset map value
		categorizedIdList[category] = catIdList
	}
}

func LoadNameMapping() (name2id map[string]string, id2name map[string]string) {
	raw, err := ioutil.ReadFile("./mappings.json")
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	var c map[string]map[string]string
	json.Unmarshal(raw, &c)
	return c["name2id"], c["id2name"]
}

// Returns empty string when nothing could be matched
func GetIdForFuzzyName(propCategory string, fuzzyName string) string {
	lowerFuzzyName := strings.ToLower(fuzzyName)
	// check if it is a valid id, no need for fuzzy lookup in that case.
	// Also, it must be within the correct category
	if x := id2name[lowerFuzzyName]; x != "" && categorizedIdList[lowerFuzzyName] != nil {
		return lowerFuzzyName
	}

	// Use a bag-size of 2, 3 and 4
	cm := closestmatch.New(categorizedNameList[propCategory], []int{2, 3, 4, 5})

	// Get the best match, and lower-case it:
	// the named2id mapping is lower-cased for less error prone matching
	bestGuess := strings.ToLower(cm.Closest(fuzzyName))

	return name2id[bestGuess]
}