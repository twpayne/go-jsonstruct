package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/pflag"

	"github.com/twpayne/go-jsonstruct/v3"
)

var (
	abbreviations            = pflag.String("abbreviations", "", "comma-separated list of extra abbreviations")
	format                   = pflag.String("format", "json", "format (json or yaml)")
	decompress               = pflag.BoolP("gzip", "z", false, "decompress input with gzip")
	fileHeader               = pflag.String("file-header", "", "file header")
	omitEmptyTags            = pflag.String("omitempty-tags", "auto", "generate ,omitempty tags (never, always, or auto)")
	packageComment           = pflag.String("package-comment", "", "package comment")
	packageName              = pflag.String("package-name", "main", "package name")
	skipUnparsableProperties = pflag.Bool("skip-unparsable-properties", true, "skip unparsable properties")
	stringTags               = pflag.Bool("string-tags", false, "generate ,string tags")
	structTagName            = pflag.String("struct-tag-name", "", "struct tag name")
	typeComment              = pflag.String("type-comment", "", "type comment")
	typeName                 = pflag.String("typename", "T", "type name")
	intType                  = pflag.String("int-type", "", "integer type")
	useJSONNumber            = pflag.Bool("use-json-number", false, "use json.Number")
	goFormat                 = pflag.Bool("go-format", true, "format generated Go code")
	output                   = pflag.StringP("output", "o", "", "output filename")

	omitEmptyTagsType = map[string]jsonstruct.OmitEmptyTagsType{
		"never":  jsonstruct.OmitEmptyTagsNever,
		"always": jsonstruct.OmitEmptyTagsAlways,
		"auto":   jsonstruct.OmitEmptyTagsAuto,
	}
)

func run() error {
	pflag.Parse()

	options := []jsonstruct.GeneratorOption{
		jsonstruct.WithFileHeader(*fileHeader),
		jsonstruct.WithOmitEmptyTags(omitEmptyTagsType[*omitEmptyTags]),
		jsonstruct.WithSkipUnparsableProperties(*skipUnparsableProperties),
		jsonstruct.WithStringTags(*stringTags),
		jsonstruct.WithUseJSONNumber(*useJSONNumber),
		jsonstruct.WithGoFormat(*goFormat),
	}
	if *abbreviations != "" {
		extraAbbreviations := strings.Split(strings.ToUpper(*abbreviations), ",")
		options = append(options, jsonstruct.WithExtraAbbreviations(extraAbbreviations...))
	}
	if *intType != "" {
		options = append(options, jsonstruct.WithIntType(*intType))
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

	if pflag.NArg() == 0 {
		var input io.Reader = os.Stdin
		if *decompress {
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
			for _, arg := range pflag.Args() {
				if err := generator.ObserveJSONFile(arg); err != nil {
					return err
				}
			}
		case "yaml":
			for _, arg := range pflag.Args() {
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
