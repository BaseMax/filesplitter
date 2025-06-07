package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/fatih/color"
)

func logInfo(msg string) {
	color.Green("✅ %s", msg)
}

func logError(msg string) {
	color.Red("❌ %s", msg)
}

func logWarn(msg string) {
	color.Yellow("⚠️  %s", msg)
}

func printBanner() {
	color.Cyan(`
📁 FileSplitter v1.0 by BaseMax
📦 Split large files by lines or size with style!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`)
}

func main() {
	printBanner()

	inputFile := flag.String("in", "", "Input file path (e.g., usernames.txt)")
	linesPerFile := flag.Int("lines", 0, "Split by number of lines (e.g., 1000000)")
	sizePerFile := flag.String("size", "", "Split by max size (e.g., 100MB, 500KB)")
	outPrefix := flag.String("prefix", "part", "Output filename prefix")
	outputDir := flag.String("outdir", ".", "Output directory")

	flag.Parse()

	if *inputFile == "" {
		logError("Input file is required! Use -in flag.")
		os.Exit(1)
	}

	file, err := os.Open(*inputFile)
	if err != nil {
		logError("Failed to open input file: " + err.Error())
		os.Exit(1)
	}
	defer file.Close()

	stat, _ := file.Stat()
	logInfo(fmt.Sprintf("📄 Input File: %s (%.2f MB)", *inputFile, float64(stat.Size())/(1024*1024)))

	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(nil)
	lineCount := 0
	part := 1
	var written int64 = 0
	var maxSizeBytes int64 = parseSize(*sizePerFile)

	var out *os.File

	createNewPart := func() {
		if out != nil {
			writer.Flush()
			out.Close()
		}
		filename := filepath.Join(*outputDir, fmt.Sprintf("%s%d.txt", *outPrefix, part))
		var err error
		out, err = os.Create(filename)
		if err != nil {
			logError("Cannot create output file: " + err.Error())
			os.Exit(1)
		}
		writer = bufio.NewWriter(out)
		logInfo(fmt.Sprintf("✂️  Creating: %s", filename))
		part++
		written = 0
		lineCount = 0
	}

	createNewPart()

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			logError("Error reading line: " + err.Error())
			break
		}

		if (*linesPerFile > 0 && lineCount >= *linesPerFile) || (maxSizeBytes > 0 && written+int64(len(line)) > maxSizeBytes) {
			createNewPart()
		}

		writer.WriteString(line)
		lineCount++
		written += int64(len(line))
	}

	writer.Flush()
	if out != nil {
		out.Close()
	}

	logInfo("🎉 Done! All parts created.")
}

func parseSize(sizeStr string) int64 {
	if sizeStr == "" {
		return 0
	}
	unit := sizeStr[len(sizeStr)-2:]
	numStr := sizeStr[:len(sizeStr)-2]

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		logWarn("Failed to parse size, ignoring size limit.")
		return 0
	}

	switch unit {
	case "KB", "kb":
		return int64(num * 1024)
	case "MB", "mb":
		return int64(num * 1024 * 1024)
	case "GB", "gb":
		return int64(num * 1024 * 1024 * 1024)
	default:
		logWarn("Unknown size unit, use KB/MB/GB.")
		return 0
	}
}
