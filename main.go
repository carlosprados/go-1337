package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
)

var leetMapping = map[string]string{
	"a": "4", "b": "8", "c": "<", "d": "|)", "e": "3", "f": "ph", "g": "6", "h": "#", "i": "1", "j": "]", "k": "|<", "l": "1", "m": "/\\/\\", "n": "|\\|", "o": "0", "p": "|*", "q": "(,)", "r": "|2", "s": "5", "t": "7", "u": "|_|", "v": "\\/", "w": "\\/\\/", "x": "><", "y": "`/", "z": "2",
}

var reverseLeetMapping = func() map[string]string {
	reverse := make(map[string]string)
	for k, v := range leetMapping {
		reverse[v] = k
	}
	return reverse
}()

func toLeet(text string) string {
	var result strings.Builder
	for _, ch := range text {
		lowerChar := strings.ToLower(string(ch))
		if leetChar, found := leetMapping[lowerChar]; found {
			result.WriteString(leetChar)
		} else {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

func fromLeet(text string) string {
	var result strings.Builder
	for i := 0; i < len(text); {
		found := false
		for k, v := range reverseLeetMapping {
			if strings.HasPrefix(text[i:], k) {
				result.WriteString(v)
				i += len(k)
				found = true
				break
			}
		}
		if !found {
			result.WriteByte(text[i])
			i++
		}
	}
	return result.String()
}

func main() {
	toLeetFlag := flag.Bool("to1337", false, "Convert text to 1337")
	fromLeetFlag := flag.Bool("from1337", false, "Convert 1337 text to normal")
	flag.Parse()

	if !*toLeetFlag && !*fromLeetFlag {
		fmt.Println("Please specify either -to1337 or -from1337 flag.")
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Enter text:")
	if scanner.Scan() {
		input := scanner.Text()
		if *toLeetFlag {
			fmt.Println("1337 Text:", toLeet(input))
		} else if *fromLeetFlag {
			fmt.Println("Normal Text:", fromLeet(input))
		}
	}
}
