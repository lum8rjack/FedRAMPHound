package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printRootUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "upload":
		os.Exit(runUpload(os.Args[2:]))
	case "collect":
		os.Exit(runCollect(os.Args[2:]))
	case "-h", "-help", "--help", "help":
		printRootUsage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command %q\n\n", os.Args[1])
		printRootUsage()
		os.Exit(1)
	}
}

func printRootUsage() {
	fmt.Fprintln(os.Stderr, `usage:
  fedramphound collect [flags]
  fedramphound upload [flags]

collect  Build an OpenGraph payload from FedRAMP Marketplace JSON dumps
upload   Upload the FedRAMP Marketplace extension schema to BloodHound`)
}

func runCollect(args []string) int {
	fs := flag.NewFlagSet("collect", flag.ExitOnError)
	fileAgencies := fs.String("agencies", "", "agencies JSON file")
	fileAssessors := fs.String("assessors", "", "assessors JSON file")
	fileProducts := fs.String("products", "", "products JSON file")
	outfile := fs.String("output", "fedramp-opengraph.json", "OpenGraph output JSON file")
	_ = fs.Parse(args)

	if *fileAgencies == "" || *fileAssessors == "" || *fileProducts == "" {
		fmt.Fprintln(os.Stderr, "usage: fedramphound collect -agencies agencies.json -assessors assessors.json -products products.json [-output fedramp-opengraph.json]")
		return 1
	}

	data, err := loadFiles(*fileAgencies, *fileAssessors, *fileProducts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading input: %v\n", err)
		return 1
	}

	graph := BuildOpenGraph(data)

	out, err := json.MarshalIndent(graph, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding OpenGraph: %v\n", err)
		return 1
	}

	if err := os.WriteFile(*outfile, out, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", *outfile, err)
		return 1
	}

	fmt.Printf("Wrote %s (%s)\n", *outfile, summarizeGraph(graph))
	return 0
}
