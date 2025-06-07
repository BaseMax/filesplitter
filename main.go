package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

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
📁 FileSplitter v2.0 by BaseMax
📦 Split massive files by lines or size with style!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`)
}

func main() {
	printBanner()

	inputFile := flag.String("in", "", "Input file path (e.g., usernames.txt)")
	linesPerFile := flag.Int("lines", 0, "Split by number of lines (e.g., 1000000)")
	sizePerFile := flag.String("size", "", "Split by max size (e.g., 100MB, 500KB)")
	outPrefix := flag.String("prefix", "part", "Output filename prefix")
	outputDir := flag.String("outdir", ".", "Output directory")
	fileExt := flag.String("ext", "txt", "Output file extension")
	padWidth := flag.Int("pad", 3, "Zero padding width for file index")
	timestamp := flag.Bool("ts", false, "Add timestamp to filenames")
	dryRun := flag.Bool("dry", false, "Dry run mode (preview only)")

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

	maxSizeBytes, err := parseSize(*sizePerFile)
	if err != nil {
		logWarn("Invalid size format: " + err.Error())
		maxSizeBytes = 0
	}

	splitFile(file, *linesPerFile, maxSizeBytes, *outputDir, *outPrefix, *fileExt, *padWidth, *timestamp, *dryRun)
}

func splitFile(file *os.File, maxLines int, maxSizeBytes int64, outputDir, prefix, ext string, padWidth int, useTS, dryRun bool) {
	reader := bufio.NewReader(file)
	lineCount := 0
	part := 1
	var written int64 = 0
	var out *os.File
	var writer *bufio.Writer

	createNewPart := func() error {
		if out != nil {
			writer.Flush()
			out.Close()
		}
		suffix := fmt.Sprintf("%0*d", padWidth, part)
		if useTS {
			suffix = fmt.Sprintf("%s_%s", suffix, time.Now().Format("20060102_150405"))
		}
		filename := filepath.Join(outputDir, fmt.Sprintf("%s%s.%s", prefix, suffix, ext))
		if dryRun {
			logInfo("[DryRun] Would create: " + filename)
			return nil
		}
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		out = f
		writer = bufio.NewWriter(out)
		logInfo("✂️  Creating: " + filename)
		written = 0
		lineCount = 0
		part++
		return nil
	}

	err := createNewPart()
	if err != nil {
		logError("Unable to start: " + err.Error())
		return
	}

	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			logError("Error reading line: " + err.Error())
			break
		}

		if (maxLines > 0 && lineCount >= maxLines) || (maxSizeBytes > 0 && written+int64(len(line)) > maxSizeBytes) {
			err := createNewPart()
			if err != nil {
				logError("Failed to create new part: " + err.Error())
				break
			}
		}

		if !dryRun {
			writer.WriteString(line)
		}
		lineCount++
		written += int64(len(line))
	}

	if !dryRun && writer != nil {
		writer.Flush()
		if out != nil {
			out.Close()
		}
	}

	logInfo("🎉 Done! All parts created.")
}

func parseSize(sizeStr string) (int64, error) {
	if sizeStr == "" {
		return 0, nil
	}
	sizeStr = strings.TrimSpace(strings.ToUpper(sizeStr))
	re := regexp.MustCompile(`(?i)^(\d+(\.\d+)?)(KB|MB|GB|B)$`)
	matches := re.FindStringSubmatch(sizeStr)
	if len(matches) != 4 {
		return 0, errors.New("invalid size format")
	}

	num, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, err
	}

	switch matches[3] {
	case "B":
		return int64(num), nil
	case "KB":
		return int64(num * 1024), nil
	case "MB":
		return int64(num * 1024 * 1024), nil
	case "GB":
		return int64(num * 1024 * 1024 * 1024), nil
	default:
		return 0, errors.New("unknown size unit")
	}
}
