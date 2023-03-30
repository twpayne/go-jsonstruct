package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/twpayne/go-jsonstruct/v2"
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
	useJSONNumber             = flag.Bool("usejsonnumber", false, "use json.Number")
	goFormat                  = flag.Bool("goformat", true, "format generated Go code")
	output                    = flag.String("o", "", "output filename")

	omitEmptyOption = map[string]jsonstruct.OmitEmptyOption{
		"never":  jsonstruct.OmitEmptyNever,
		"always": jsonstruct.OmitEmptyAlways,
		"auto":   jsonstruct.OmitEmptyAuto,
	}
)

func run() error {
	flag.Parse()

	options := []jsonstruct.GeneratorOption{
		jsonstruct.WithOmitEmpty(omitEmptyOption[*omitempty]),
		jsonstruct.WithSkipUnparseableProperties(*skipUnparseableProperties),
		jsonstruct.WithUseJSONNumber(*useJSONNumber),
		jsonstruct.WithGoFormat(*goFormat),
	}
	if *abbreviations != "" {
		options = append(options, jsonstruct.WithExtraAbbreviations(strings.Split(*abbreviations, ",")...))
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
	if *format == "yaml" && *structTagName == "" {
		*structTagName = "yaml"
	}
	if *structTagName != "" {
		options = append(options, jsonstruct.WithStructTagName(*structTagName))
	}

	generator := jsonstruct.NewGenerator(options...)

	if flag.NArg() == 0 {
		var input io.Reader = os.Stdin
		if *uncompress {
			var err error
			input, err = gzip.NewReader(input)
			if err != nil {
				return err
			}
		}

		switch *format {
		case "json":
			if err := generator.ObserveJSONReader(input); err != nil {
				return err
			}
		case "yaml":
			if err := generator.ObserveYAMLReader(input); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown format: %s", *format)
		}
	} else {
		switch *format {
		case "json":
			for _, arg := range flag.Args() {
				if err := generator.ObserveJSONFile(arg); err != nil {
					return err
				}
			}
		case "yaml":
			for _, arg := range flag.Args() {
				if err := generator.ObserveYAMLFile(arg); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unknown format: %s", *format)
		}
	}

	goCode, err := generator.Generate()
	if err != nil {
		return err
	}

	if *output == "" || *output == "-" {
		_, err = os.Stdout.Write(goCode)
		return err
	}

	return os.WriteFile(*output, goCode, 0o666) //nolint:gosec
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
