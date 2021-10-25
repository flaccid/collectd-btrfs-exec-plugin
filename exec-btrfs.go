package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/urfave/cli"
)

func findSubstring(word string, list []string) (int, bool) {
	for i, v := range list {
		if strings.Index(v, word) >= 0 {
			return i, true
		}
	}
	return -1, false
}

func putVal(hostname string, fsName string, metric string,
	interval int, value string) {
	fmt.Printf("PUTVAL %v/exec-btrfs-%v/gauge-%v interval=%v N:%v\n",
		hostname, fsName, metric, interval, value)
}

func getSupportedMetrics() []string {
	metrics := []string{
		"device_unallocated",
		"free_estimated",
		"data_ratio",
		"used",
		"global_reserve",
		"inodes",
		"reserve_used",
		"device_allocated",
		"metadata_ratio",
		"device_size",
		"device_missing",
		"free_estimated_minimum",
	}
	return metrics
}

func btrfsStats(m string) map[string]interface{} {
	// sane paths
	os.Setenv("PATH",
		os.Getenv("PATH")+":/usr/bin:/usr/local/bin:/opt/local/bin:/sbin:/usr/sbin")

	// currently we'll return NaN for all if we can't read the mountpoint
	if _, err := os.Open(m); err != nil {
		metrics := getSupportedMetrics()
		btrfs := make(map[string]interface{})
		for _, e := range metrics {
			btrfs[e] = "NaN"
		}
		return btrfs
	}

	btrfsPath, err := exec.LookPath("btrfs")
	if err != nil {
		log.Fatal("cannot find the btrfs command in a common path")
	}
	btrfsArgs := []string{"fi", "usage", "--raw", m}

	var out bytes.Buffer
	out.Reset()

	cmd := exec.Command(btrfsPath, btrfsArgs...)
	cmd.Stdout = &out
	err = cmd.Run()

	if err != nil {
		log.Fatal(err)
		fmt.Println("btrfs execution failed")
	}

	btrfsLines := strings.Split(out.String(), "\n")

	/*
					  -- Example haystack --
		        Device size:		        1073741824
		        Device allocated:		         132513792
		        Device unallocated:		         941228032
		        Device missing:		                 0
		        Used:			            262144
		        Free (estimated):		         949616640	(min: 479002624)
		        Data ratio:			              1.00
		        Metadata ratio:		              2.00
		        Global reserve:		          16777216	(used: 0)
	*/

	btrfsHay := make(map[string]string)
	btrfsHay["device_size"] = "Device size:"
	btrfsHay["device_allocated"] = "Device allocated:"
	btrfsHay["device_unallocated"] = "Device unallocated:"
	btrfsHay["device_missing"] = "Device missing:"
	btrfsHay["used"] = "Used:"
	btrfsHay["free_estimated"] = "Free (estimated):"
	btrfsHay["free_estimated_minimum"] = "Free (estimated):"
	btrfsHay["data_ratio"] = "Data ratio:"
	btrfsHay["metadata_ratio"] = "Metadata ratio:"
	btrfsHay["global_reserve"] = "Global reserve:"
	btrfsHay["global_reserve_used"] = "Global reserve:"

	btrfs := make(map[string]interface{})

	for metric, needle := range btrfsHay {
		i, _ := findSubstring(needle, btrfsLines)

		line := btrfsLines[i]

		// these are 3 column exceptions to process
		switch metric {
		case "free_estimated":
			freeEstimated := strings.TrimSpace(
				strings.Split(strings.Split(line, "):")[1], "(")[0])
			btrfs["free_estimated"] = freeEstimated
		case "free_estimated_minimum":
			freeEstimatedMin := strings.TrimSpace(
				strings.Replace(strings.Split(line, ":")[2], ")", "", -1))
			btrfs["free_estimated_minimum"] = freeEstimatedMin
		case "global_reserve":
			globalReserve := strings.TrimSpace(
				strings.Split(strings.Split(line, ":")[1], "(")[0])
			btrfs["global_reserve"] = globalReserve
		case "global_reserve_used":
			globalReserveUsed := strings.TrimSpace(
				strings.Replace(strings.Split(line, ":")[2], ")", "", -1))
			btrfs["global_reserve_used"] = globalReserveUsed
		default:
			right := strings.Split(line, ":")

			// get the final right column which should contain the human value
			btrfs[metric] = strings.TrimSpace(right[len(right)-1])

			// debug only
			// fmt.Printf("debug: [%v] [%v] [%v] [%v]\n\n", metric, needle, i, line)
		}

		// do you get this?
		btrfs["inodes"] = "0"
	}

	return btrfs
}

func main() {
	localHostname, _ := os.Hostname()
	defaultInterval := 20

	app := cli.NewApp()
	app.Name = "exec-btrfs"
	app.Version = "v0.0.1"
	app.Compiled = time.Now()
	app.Authors = []*cli.Author{
		&cli.Author{
			Name:  "Chris Fordham",
			Email: "chris@fordham-nagy.id.au",
		},
	}
	app.Copyright = "(c) 2016 Chris Fordham"
	app.Usage = "Btrfs exec plugin for collectd emmitting Btrfs filesystem stats"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "hostname, H",
			Value: localHostname,
		},
		&cli.IntFlag{
			Name:  "interval, i",
			Value: defaultInterval,
		},
	}
	app.Action = func(c *cli.Context) error {
		if c.NArg() < 1 {
			fmt.Println("Usage: exec-btrfs [global options] <mountpoint>")
			os.Exit(1)
		}
		mountPoint := c.Args().Get(0)
		fsName := strings.Replace(mountPoint, "/", "-", -1)[1:]
		if (mountPoint == "/") {
			fsName = "root"
		}

		hostname := c.String("hostname")
		if (hostname == localHostname) {
			envCollecdHostname := os.Getenv("COLLECTD_HOSTNAME")
			if len(envCollecdHostname) > 0 {
				hostname = envCollecdHostname
			}
		}

		interval := c.Int("interval")
		envCollectdInterval := os.Getenv("COLLECTD_INTERVAL")
		if len(envCollectdInterval) > 0 {
			_, err := fmt.Sscan(envCollectdInterval, &interval)
			if err != nil {
				interval = c.Int("interval")
			}
		}

		// main output loop
		for {
			btrfs := btrfsStats(mountPoint)

			// debug only
			// fmt.Println("map:", btrfs)

			for metric, value := range btrfs {
				putVal(hostname,
					fsName, string(metric),
					interval,
					value.(string))
			}

			time.Sleep((time.Duration(c.Int("interval")) * 1000) * time.Millisecond)
		}
	}

	app.Run(os.Args)
}
