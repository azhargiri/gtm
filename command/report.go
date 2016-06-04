package command

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"

	"edgeg.io/gtm/report"
	"edgeg.io/gtm/scm"
	"edgeg.io/gtm/util"
	"github.com/mitchellh/cli"
)

type ReportCmd struct {
}

func NewReport() (cli.Command, error) {
	return ReportCmd{}, nil
}

func (r ReportCmd) Help() string {
	return `
	Report on time logged

	Show report for a specific sha1 commits:
	gtm report sha1 ...

	Show report by piping output from git log:
	git report -1 --pretty=%H|gtm report
	`
}

func (r ReportCmd) Run(args []string) int {
	reportFlags := flag.NewFlagSet("report", flag.ExitOnError)
	format := reportFlags.String(
		"format",
		"commits",
		"Specify report format [commits|totals|files|timeline]")
	limit := reportFlags.Int(
		"n",
		1,
		fmt.Sprintf("Limit number of log enteries"))
	totalOnly := reportFlags.Bool(
		"total-only",
		false,
		"Only display total time")
	if err := reportFlags.Parse(os.Args[2:]); err != nil {
		fmt.Println(err)
		return 1
	}

	if !util.StringInSlice([]string{"commits", "timeline", "files"}, *format) {
		fmt.Printf("report --format=%s not valid\n", *format)
		return 1
	}

	var (
		commits []string
		out     string
		err     error
	)

	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			commits = append(commits, scanner.Text())
		}
	} else {
		if len(reportFlags.Args()) == 0 {
			commits, err = scm.CommitIDs(*limit)
			if err != nil {
				fmt.Println(err)
				return 1
			}
		}
		for _, a := range reportFlags.Args() {
			if match, err := regexp.MatchString("[-|.|,|:|*]", a); err != nil || match {
				fmt.Printf("\nNot a valid commit sha1 %s\n", a)
				return 1
			}
			commits = append(commits, a)
		}
	}

	if len(commits) == 0 {
		return 0
	}

	switch *format {
	case "commits":
		out, err = report.Commits(commits, *totalOnly)
	case "files":
		out, err = report.Files(commits)
	case "timeline":
		out, err = report.Timeline(commits)
	}

	if err != nil {
		fmt.Println(err)
		return 1
	}
	fmt.Printf(out)

	return 0
}

func (r ReportCmd) Synopsis() string {
	return `
	Show commit time logs
	`
}
