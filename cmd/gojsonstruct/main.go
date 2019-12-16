package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/twpayne/go-jsonstruct"
)

var (
	abbreviations             = flag.String("abbreviations", "", "comma-separated list of extra abbreviations")
	uncompress                = flag.Bool("z", false, "decompress input with gzip")
	omitempty                 = flag.String("omitempty", "auto", "generate omitempty (never, always, or auto)")
	packageComment            = flag.String("packagecomment", "", "package comment")
	packageName               = flag.String("packagename", "main", "package name")
	skipUnparseableProperties = flag.Bool("skipunparseableproperties", true, "skip unparseable properties")
	typeComment               = flag.String("typecomment", "", "type comment")
	typeName                  = flag.String("typename", "T", "type name")

	omitEmptyOption = map[string]jsonstruct.OmitEmptyOption{
		"never":  jsonstruct.OmitEmptyNever,
		"always": jsonstruct.OmitEmptyAlways,
		"auto":   jsonstruct.OmitEmptyAuto,
	}
)

func run() error {
	flag.Parse()

	var input io.Reader = os.Stdin
	if *uncompress {
		var err error
		input, err = gzip.NewReader(input)
		if err != nil {
			return err
		}
	}

	observedValue, err := jsonstruct.Observe(input)
	if err != nil {
		return err
	}

	options := []jsonstruct.GeneratorOption{
		jsonstruct.WithOmitEmpty(omitEmptyOption[*omitempty]),
		jsonstruct.WithSkipUnparseableProperties(*skipUnparseableProperties),
	}
	if *abbreviations != "" {
		abbreviationsMap := make(map[string]bool)
		for abbreviation := range jsonstruct.WellKnownAbbreviations {
			abbreviationsMap[abbreviation] = true
		}
		for _, abbreviation := range strings.Split(*abbreviations, ",") {
			abbreviationsMap[abbreviation] = true
		}
		options = append(options, jsonstruct.WithFieldNamer(
			&jsonstruct.AbbreviationHandlingFieldNamer{
				Abbreviations: abbreviationsMap,
			},
		))
	}
	if *packageComment != "" {
		options = append(options, jsonstruct.WithPackageComment(*packageComment))
	}
	if *packageName != "" {
		options = append(options, jsonstruct.WithPackageName(*packageName))
	}
	if *typeComment != "" {
		options = append(options, jsonstruct.WithTypeComment(*typeComment))
	}
	if *typeName != "" {
		options = append(options, jsonstruct.WithTypeName(*typeName))
	}

	goCode, err := jsonstruct.NewGenerator(options...).GoCode(observedValue)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(goCode)
	return err
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
