package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
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

func btrfsStats(m string) map[string]int64 {
	btrfsPath := "/usr/bin/btrfs"
	btrfsArgs := []string{"fi", "usage", "--raw", m}

	var out bytes.Buffer
	out.Reset()

	cmd := exec.Command(btrfsPath, btrfsArgs...)
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		log.Fatal(err)
		fmt.Println("btrfs execution failed")
	}

	btrfsLines := strings.Split(out.String(), "\n")

	/*
	  -- Example haystack --
	  Device size:		   1.00GiB
	  Device allocated:		 126.38MiB
	  Device unallocated:		 897.62MiB
	  Device missing:		     0.00B
	  Used:			 256.00KiB
	  Free (estimated):		 905.62MiB	(min: 456.81MiB)
	  Data ratio:			      1.00
	  Metadata ratio:		      2.00
	  Global reserve:		  16.00MiB	(used: 0.00B)
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

	btrfs := make(map[string]int64)

	for metric, needle := range btrfsHay {
		i, _ := findSubstring(needle, btrfsLines)

		line := btrfsLines[i]

		// these are 3 column exceptions to process
		switch metric {
		case "free_estimated":
			freeEstimated, _ := strconv.ParseInt(strings.TrimSpace(strings.Split(strings.Split(line, "):")[1], "(")[0]), 10, 64)
			btrfs["free_estimated"] = freeEstimated
		case "free_estimated_minimum":
			freeEstimatedMin, _ := strconv.ParseInt(strings.TrimSpace(strings.Replace(strings.Split(line, ":")[2], ")", "", -1)), 10, 64)
			btrfs["free_estimated_minimum"] = freeEstimatedMin
		case "global_reserve":
			globalReserve, _ := strconv.ParseInt(strings.TrimSpace(strings.Split(strings.Split(line, ":")[1], "(")[0]), 10, 64)
			btrfs["global_reserve"] = globalReserve
		case "global_reserve_used":
			globalReserveUsed, _ := strconv.ParseInt(strings.TrimSpace(strings.Replace(strings.Split(line, ":")[2], ")", "", -1)), 10, 64)
			btrfs["global_reserve_used"] = globalReserveUsed
		default:
			right := strings.Split(line, ":")

			// get the final right column which should contain the human metric
			initialValue := strings.TrimSpace(right[len(right)-1])

			// oh no, its got [i]B
			if strings.Contains(initialValue, "B") {
				totalBytes := initialValue

				// we round and assign as integer assuming that the float reported
				// from btrfs is always 2 floating points so can be innacurate, thus
				// potential for fractions of a byte!
				v, _ := strconv.ParseInt(totalBytes, 10, 64)

				btrfs[metric] = v
			} else {
				metricValue, _ := strconv.ParseFloat(initialValue, 64)
				btrfs[metric] = int64(metricValue)
			}
			
			// debug only
			// fmt.Printf("*************** %v %v %v %v %v\n\n", metric, needle, i, found, line)
		}

		// do you get this?
		btrfs["inodes"] = int64(0)
	}

	return btrfs
}

func main() {
	app := cli.NewApp()
	app.Name = "exec-btrfs"
	app.Version = "v0.0.1"
	app.Compiled = time.Now()
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Chris Fordham",
			Email: "chris@fordham-nagy.id.au",
		},
	}
	app.Copyright = "(c) 2016 Chris Fordham"
	app.Usage = "BTRFS exec plugin for collectd ommitting btrfs stats"
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "hostname, H"},
		cli.StringFlag{Name: "interval, i"},
	}

	app.Action = func(c *cli.Context) error {
		hostname, _ := os.Hostname()
		interval := 20

		mountPoint := c.Args().Get(0)
		mountPointSplit := strings.Split(mountPoint, "/")
		fsName := mountPointSplit[len(mountPointSplit)-1]

		// main output loop
		for {
			btrfs := btrfsStats(mountPoint)

			// debug only
			// fmt.Println("map:", btrfs)

			for metric, value := range btrfs {
				fmt.Printf("PUTVAL %v/exec-btrfs_%v/gauge-%v interval=%v N:%v\n", hostname, fsName, metric, interval, value)
			}

			time.Sleep((time.Duration(interval) * 1000) * time.Millisecond)
		}
	}

	app.Run(os.Args)
}
