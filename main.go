package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	var inputFileName, outputFileName string
	args := os.Args[1:]
	if len(args) == 1 {
		inputFileName = args[0]
		outputFileName = genOutputFileName(inputFileName)
	} else if len(args) == 2 {
		inputFileName = args[0]
		outputFileName = args[1]
	} else {
		fmt.Println("Uh-Oh! You've entered too many arguments")
		fmt.Println()
		fmt.Println("How to Use:")
		fmt.Println("./csvToDescription <inputCSVFileLocation>")
		fmt.Println("                          OR                                     ")
		fmt.Println("./csvToDescription <inputCSVFileLocation> <outputTXTFileLocation>")
		fmt.Println()
		fmt.Println("If you omit the 2nd argument for the output file location the program")
		fmt.Println("will automatically generate a new file in the same directory as this")
		fmt.Println("program.")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("./csvToDescription ~/Documents/podcasts/see_the_thing_is.csv")
		fmt.Println("                          OR                                ")
		fmt.Println("./csvToDescription ~/Documents/podcasts/see_the_thing_is.csv ~/Desktop/see_the_thing_is_converted.txt")
		os.Exit(0)
	}
	runWithScanner(inputFileName, outputFileName)
	fmt.Printf("Given the input file: %s\n", inputFileName)
	fmt.Printf("This is the converted file: %s\n", outputFileName)
}

func runWithScanner(inputFileName, outputFileName string) {
	scanr, closer, err := createScannerAndLoadWithHeaders(inputFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer closer()
	outputWriter, outputCloser, outputErr := createOutputFile(outputFileName)
	if outputErr != nil {
		log.Fatal(outputErr)
	}
	defer outputCloser()
	nameIdx, startIdx := getNameAndStartIndexes(scanr.Bytes())
	if nameIdx == -1 || startIdx == -1 {
		log.Fatalf("name=%d, start=%d", nameIdx, startIdx)
	}
	err = exhaustScanner(scanr, outputWriter, nameIdx, startIdx)
	if err != nil {
		log.Fatal(err)
	}
}

func exhaustScanner(s *bufio.Scanner, outputFile *os.File, nameIdx, startIdx int) error {
	var (
		record        [][]byte
		formattedLine []byte
		written       int
		writeErr      error
	)
	for s.Scan() {
		record = parseLine(s.Bytes())
		formattedLine = formatLine(record[nameIdx], record[startIdx])
		written, writeErr = outputFile.Write(formattedLine)
		if writeErr != nil {
			return errors.New(fmt.Sprintf("Failed to write %d bytes %e", written, writeErr))
		}
		written, writeErr = outputFile.Write([]byte{'\n'})
		if writeErr != nil {
			return errors.New(fmt.Sprintf("Failed to write newLine %d bytes %e", written, writeErr))
		}
	}
	return nil
}

func genOutputFileName(inputPath string) string {
	dir, file := filepath.Split(inputPath)
	actualName := strings.Split(file, ".")[0]
	return fmt.Sprintf("%s%s_converted-%s.txt", dir, actualName, time.Now().Format(time.RFC3339))
}

func formatLine(name, start []byte) []byte {
	return bytes.Join([][]byte{bytes.Split(start, []byte{'.'})[0], name}, []byte{'\t'})
}

func getNameAndStartIndexes(headers []byte) (int, int) {
	columnNamesBytes := parseLine(headers)
	return findIndexOfString(columnNamesBytes, "Name"), findIndexOfString(columnNamesBytes, "Start")
}

func findIndexOfString(strs [][]byte, match string) int {
	for i, str := range strs {
		if strings.Contains(string(str), match) {
			return i
		}
	}
	return -1
}

func parseLine(line []byte) [][]byte {
	return bytes.Split(line, []byte{'\t'})
}

func createScannerAndLoadWithHeaders(path string) (*bufio.Scanner, func(), error) {
	csvFile, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}

	deferFn := func() {
		if err = csvFile.Close(); err != nil {
			log.Fatal(err)
		}
	}
	s := bufio.NewScanner(csvFile)
	s.Scan()

	return s, deferFn, nil
}

func createOutputFile(path string) (*os.File, func(), error) {
	txtFile, err := os.Create(path)
	if err != nil {
		return nil, nil, err
	}

	deferFn := func() {
		if err = txtFile.Close(); err != nil {
			log.Fatal(err)
		}
	}

	return txtFile, deferFn, nil
}
