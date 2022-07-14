package main

import (
	//  "errors"
	//  "flag"
	"fmt"
	// "io"
	//  "io/ioutil"
	"bufio"
	"encoding/csv"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"strings"
	"time"
)

var TAGS = "CSV2LINE_TAGS"
var FIELDS = "CSV2LINE_FIELDS"

//----------

func readCsv(inputFile string) (output []map[string]string, err error) {

	f, err := os.Open(inputFile)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	//----- Read

	var b strings.Builder

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		t := scanner.Text()

		// Remove the comment lines.
		if len(t) > 0 && t[0] != '#' {
			b.WriteString(t)
			b.WriteString("\n")
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading error:", err)
	}

	r := csv.NewReader(strings.NewReader(b.String()))
	records, err2 := r.ReadAll()
	if err2 != nil {
		log.Fatal(err2)
	}

	//----- Format

	output = []map[string]string{}
	header := records[0]
	for i := 1; i < len(records); i++ {
		dict := map[string]string{}
		for j := range header {
			dict[header[j]] = records[i][j]
		}
		output = append(output, dict)
	}
	return output, nil
}

func writeCsv(outputFile string, records []map[string]string) {
	f, err := os.OpenFile(outputFile, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0644)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	var tags = strings.Split(os.ExpandEnv("${"+TAGS+"}"), ",")
	fmt.Printf("%q: %q\n", TAGS, tags)
	var fields = strings.Split(os.ExpandEnv("${"+FIELDS+"}"), ",")
	fmt.Printf("%q: %q\n", FIELDS, fields)

	// startTime := time.Now().Unix() - 10 ^ 6
	writer := bufio.NewWriter(f)
	for i := 0; i < len(records); i++ {
		writer.WriteString(records[i]["_measurement"])
		for _, tag := range tags {
			writer.WriteString(fmt.Sprintf(",%v=\"%v\"", tag, records[i][tag]))
		}
		writer.WriteString(" ")
		for k, field := range fields {
			if k == 0 {
				writer.WriteString(fmt.Sprintf("%v=\"%v\"", field, records[i][field]))
			} else {
				writer.WriteString(fmt.Sprintf(",%v=\"%v\"", field, records[i][field]))
			}
		}
		// writer.WriteString(fmt.Sprintf(" %d", startTime+int64(i)))
		t, err := time.Parse(time.RFC3339Nano, records[i]["_time"])
		if err != nil {
			log.Fatal(err)
		}
		writer.WriteString(fmt.Sprintf(" %d", t.UnixNano()))
		writer.WriteString("\n")
	}
	writer.Flush()
}

func main() {

	app := &cli.App{}
	// app.UseShortOptionHandling = true

	app.Action = func(c *cli.Context) error {
		input := c.Args().Get(0)
		output := c.Args().Get(1)
		fmt.Printf("Input: %q, Output: %q\n", input, output)

		records, _ := readCsv(input)
		writeCsv(output, records)
		return nil
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
