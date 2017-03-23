package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/spf13/cobra"
)

// The Talk Data Structure
type Talk struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Desc   string `json:"desc"`
	Hidden bool   `json:"hidden"`
}

// TalkArr an array of talks
type TalkArr []Talk

// Implementing functions for purpose of sorting talks by type
func (a TalkArr) Len() int           { return len(a) }
func (a TalkArr) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TalkArr) Less(i, j int) bool { return a[i].Type < a[j].Type }

func init() {
	getTalksCmd.Flags().StringVarP(&flagTalkType, "type", "t", "", "Type of a talk. Leave empty to display all")

	postTalkCmd.Flags().StringVarP(&flagTalkName, "name", "n", "", "The presenter of the talk")
	postTalkCmd.Flags().StringVarP(&flagTalkType, "type", "t", "", "Type of a talk")
	postTalkCmd.Flags().StringVarP(&flagTalkDesc, "desc", "d", "", "Description of a talk")

	RootCmd.AddCommand(getTalksCmd)
	RootCmd.AddCommand(postTalkCmd)
}

var flagTalkName, flagTalkType, flagTalkDesc string

var RootCmd = &cobra.Command{
	Use:   "talks-cli",
	Short: "Prints help for TalksCLI",
}

var getTalksCmd = &cobra.Command{
	Use:   "talks",
	Short: "Print visible talks",
	Long:  "Prints any talk in the talks database that has the hidden flag off",
	Run: func(cmd *cobra.Command, args []string) {
		getTalks(args)
	},
}

var postTalkCmd = &cobra.Command{
	Use:   "new",
	Short: "Create new talk",
	Long:  "Creates a new talk with three following arguments of name, type, and description. Do note talk submission is only allowed on COSI's subnets",
	Run: func(cmd *cobra.Command, args []string) {
		postTalk(args)
	},
}

func getTalks(args []string) {
	resp, err := http.Get("http://talks.cosi.clarkson.edu/api/talks/visible")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	talkArr := formatResponse(resp)

	// Filter talks if flag is present, sort by talks otherwise
	if len(flagTalkType) != 0 {
		talkArr = filterTalksByType(talkArr, flagTalkType)
	} else {
		sort.Sort(TalkArr(talkArr))
	}

	fmt.Printf("There are currently %d talks scheduled\n", talkArr.Len())

	for _, v := range talkArr {
		fmt.Printf("[%s]: %s by %s\n", v.Type, v.Desc, v.Name)
	}
}

func postTalk(args []string) {
	if len(flagTalkName) == 0 || len(flagTalkType) == 0 || len(flagTalkDesc) == 0 {
		fmt.Println("Error: One of the three fields is null. Exiting...")
		os.Exit(-1)
	}

	talk := Talk{
		Name: flagTalkName,
		Type: flagTalkType,
		Desc: flagTalkDesc,
	}

	buff := new(bytes.Buffer)

	json.NewEncoder(buff).Encode(talk)

	resp, err := http.Post("http://talks.cosi.clarkson.edu/api/postTalk", "application/json; charset=utf-8", buff)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
}

func formatResponse(resp *http.Response) TalkArr {
	talkArr := TalkArr{}

	err := json.NewDecoder(resp.Body).Decode(&talkArr)
	if err != nil {
		panic(err)
	}

	return talkArr
}

func filterTalksByType(tArr TalkArr, ty string) (ret TalkArr) {
	for _, t := range tArr {
		if t.Type == ty {
			ret = append(ret, t)
		}
	}
	return
}
