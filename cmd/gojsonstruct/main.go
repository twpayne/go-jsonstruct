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
	format                    = flag.String("format", "json", "format (json or yaml)")
	uncompress                = flag.Bool("z", false, "decompress input with gzip")
	omitempty                 = flag.String("omitempty", "auto", "generate omitempty (never, always, or auto)")
	packageComment            = flag.String("packagecomment", "", "package comment")
	packageName               = flag.String("packagename", "main", "package name")
	skipUnparseableProperties = flag.Bool("skipunparseableproperties", true, "skip unparseable properties")
	structTagName             = flag.String("structtagname", "", "struct tag name")
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

	var observedValue *jsonstruct.ObservedValue
	var err error
	switch *format {
	case "json":
		observedValue, err = jsonstruct.ObserveJSON(input)
	case "yaml":
		observedValue, err = jsonstruct.ObserveYAML(input)
	default:
		return fmt.Errorf("unknown format: %s", *format)
	}
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
	if *structTagName != "" {
		options = append(options, jsonstruct.WithStructTagName(*structTagName))
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
