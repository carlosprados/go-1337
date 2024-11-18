# Leet Translator CLI

A simple command-line tool written in Go that can convert text to and from `1337` (leet) speak. This project provides a way to have fun with text transformations, translating normal text into `1337` or converting `1337` text back to readable form.

## Features

- Convert any text to `1337` speak using a simple command.
- Convert `1337` speak back to normal text.
- Designed with an extensible character mapping that allows multiple characters for a single letter.

## Usage

### Building the Project

To build the project, first ensure you have Go installed. Then, run the following command:

```bash
go build -o leet_converter main.go
```

This command will generate an executable binary named `leet_converter` in your current directory.

### Running the Program

The program offers two main functions: converting text to `1337` speak and converting `1337` back to normal text.

To convert text to `1337` speak, use the `-to1337` flag:

```bash
./leet_converter -to1337
```

To convert `1337` speak back to normal text, use the `-from1337` flag:

```bash
./leet_converter -from1337
```

After running the command, the program will prompt you to enter the text to be converted.

## Example

```bash
$ ./leet_converter -to1337
Enter text:
Hello World
1337 Text: #3110 /\/0r1d
```

```bash
$ ./leet_converter -from1337
Enter text:
#3110 /\/0r1d
Normal Text: hello world
```

## Dependencies

This project has no external dependencies beyond the Go standard library. Ensure you have Go installed to compile and run the program.

## Modifying the Leet Mapping

The `leetMapping` variable can be modified to include additional mappings or change existing ones. Each character is mapped to its corresponding `1337` representation. To modify the mapping, edit the `leetMapping` in `leet_converter.go` as follows:

```go
var leetMapping = map[string]string{
    "a": "4", "b": "8", // Add or modify mappings here
    // ...
}
```

The program also has a reverse mapping (`reverseLeetMapping`) that is dynamically generated, allowing conversions from `1337` back to normal text.

## Contribution

Feel free to contribute to this project by opening issues or submitting pull requests. Suggestions for expanding the `leet` character mappings are welcome!

## License

This project is licensed under the MIT License. See the `LICENSE` file for more details.
