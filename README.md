# go-jsonstruct

[![PkgGoDev](https://pkg.go.dev/badge/github.com/twpayne/go-jsonstruct/v2)](https://pkg.go.dev/github.com/twpayne/go-jsonstruct/v2)

Generate Go structs from multiple JSON objects.

* [What does go-jsonstruct do and why should I use it?](#what-does-go-jsonstruct-do-and-why-should-i-use-it)
* [How do I use go-jsonstruct?](#how-do-i-use-go-jsonstruct)
* [What are go-jsonstruct's key features?](#what-are-go-jsonstructs-key-features)
* [How does go-jsonstruct work?](#how-does-go-jsonstruct-work)
* [License](#license)

## What does go-jsonstruct do and why should I use it?

go-jsonstruct generates Go structs from *multiple* JSON objects. Existing Go
struct generators such as [json-to-go](https://mholt.github.io/json-to-go/) and
[json2struct](http://json2struct.mervine.net/) take only a *single* JSON object
as input. go-jsonstruct takes multiple JSON objects as input and generates the
most specific Go struct possible into which all the input objects can be
unmarshalled.

This is useful if you have a collection of JSON objects, where no single object
has all properties present, and you want to unmarshal those JSON objects into a
Go program. Example collections include:

* JSON responses received from a REST API with no defined schema.
* Multiple values from a JSON column in an SQL database.
* All the JSON documents in a document database.

## How do I use go-jsonstruct?

Install go-jsonstruct:

```sh
go install github.com/twpayne/go-jsonstruct/v2/cmd/gojsonstruct@latest
```

Feed it some JSON objects. For example you can feed it with

```json
{
  "age": 37,
  "user_height_m": 2
}

{
  "age": 38,
  "user_height_m": 1.7,
  "favoriteFoods": [
    "cake"
  ]
}
```

by running

```sh
echo '{"age":37,"user_height_m":2}' \
    '{"age":38,"user_height_m":1.7,"favoriteFoods":["cake"]}' \
    | gojsonstruct
```

This will output:

```go
package main

type T struct {
    Age           int      `json:"age"`
    FavoriteFoods []string `json:"favoriteFoods,omitempty"`
    UserHeightM   float64  `json:"user_height_m"`
}
```

This example demonstrates:

* `age` is always observed as an integer, and so is given type `int`. The
  lower-case JSON property `age` is converted into an exported Go field name
  `Age` for compatibility with `encoding/json`.
* `favoriteFoods` is observed as a JSON array, but is not always present, but
  when it is present it only contains JSON strings, and so is given type
  `[]string`. The `camelCase` name `favoriteFoods` is converted into the
  exported Go field name `FavoriteFoods` and is tagged with `omitempty`.
* `user_height_m` is observed as JSON numbers `2` and `1.7`, for which the most
  specific Go type is `float64`. The `snake_case` name `user_height_m` is
  converted to the exported Go field name `UserHeightM`.
* Properties are sorted alphabetically.

go-jsonstruct recursively handles nested array elements and objects. For
example, given the following three JSON objects input:

```json
{
  "nested": {
    "bar": true,
    "foo": "baz"
  }
}

{
  "nested": {
    "bar": false,
    "foo": null
  }
}

{
  "nested": {
    "bar": true,
    "foo": ""
  }
}
```

which you can try by running

```sh
echo '{"nested":{"bar":true,"foo":"baz"}}' \
    '{"nested":{"bar":false,"foo":null}}' \
    '{"nested":{"bar":true,"foo":""}}' \
    | gojsonstruct --package-name mypackage --typename MyType
```

generates the output

```go
package mypackage

type MyType struct {
    Nested struct {
        Bar bool    `json:"bar"`
        Foo *string `json:"foo"`
    } `json:"nested"`
}
```

This demonstrates:

* The package name and type name can be set on the command line.
* The arbitrarily-complex property `nested` is recursively converted to a nested
  `struct` that all values can be unmarshalled to. go-jsonstruct handles array
  elements in an identical fashion, resolving array elements to the
  most-specific type.
* `nested.bar` is always observed as a JSON bool, and is converted to a field of
  type `bool`.
* `nested.foo` is observed as a JSON string, a JSON null, and an empty JSON
  string and is converted to a field of type `*string` without `omitempty`. With
  `omitempty`, Go's `encoding/json` omits the field in the two cases of  `nil`
  and a pointer to an empty `string`, so there is no way to distinguish between
  the observed values of `null` and `""`. go-jsonstruct falls back to the option
  of `*string` without `omitempty` which means that a value is always present,
  even if empty.

You can feed it your own data via the standard input, for example if you have a
file with one JSON object per line in `objects.json` you can run:

    gojsonstruct < objects.json

To learn about more about the available options, run:

    gojsonstruct --help

## What are go-jsonstruct's key features?

* Finds the most specific Go type that can represent all input values.
* Generates Go struct field names from  `camelCase`, `kebab-case`, and
  `snake_case` JSON object property names.
* Capitalizes common abbreviations (e.g. HTTP, ID, and URL) when
  generating Go struct field names to follow Go conventions, with the option to
  add your own abbreviations.
* Gives you control over the output, including the generated package name, type
  name, and godoc-compatible comments.
* Generates deterministic output based only on the determined structure of the
  input, making it suitable for incorporation into build pipelines or detecting
  schema changes.
* Generates `,omitempty` tags.
* Generates `,string` tags.
* Uses the standard library's `time.Time` when possible.
* Gracefully handles properties with spaces that [cannot be unmarshalled by
  `encoding/json`](https://github.com/golang/go/issues/18531).

## How does go-jsonstruct work?

go-jsonstruct consists of two phases: observation and code generation.

Firstly, in the observation phase, go-jsonstruct explores all the input objects
and records statistics on what types are observed in each part. It recurses into
objects and iterates over arrays.

Secondly, in the code generation phase, go-jsonstruct inspects the gathered
statistics and determines the strictest possible Go type that can represent all
the observed values. For example, the values `0` and `1` can be represented as
an `int`, the values `0`, `1`, and `2.2` require a `float64`, and `true`, `3.3`,
and `"name"` require an `any`.

## License

BSD
