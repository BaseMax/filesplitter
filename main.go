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

const bufSize = 128 * 1024 // 128KB buffer for I/O

func logInfo(msg string)    { color.Green("✅ %s", msg) }
func logError(msg string)   { color.Red("❌ %s", msg) }
func logWarn(msg string)    { color.Yellow("⚠️  %s", msg) }
func logSuccess(msg string) { color.Cyan("🎯 %s", msg) }

func printBanner() {
	color.Cyan(`
📁 FileSplitter v1.0 by Max Base
📦 Split massive files by lines, size, or pattern with style!
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
`)
}

func main() {
	printBanner()

	inputFile := flag.String("in", "", "Input file path (e.g., usernames.txt)")
	linesPerFile := flag.Int("lines", 0, "Split by number of lines (e.g., 1000000)")
	sizePerFile := flag.String("size", "", "Split by max size (e.g., 100MB, 500KB)")
	pattern := flag.String("pattern", "", "Split file whenever this pattern is matched")
	outPrefix := flag.String("prefix", "part", "Output filename prefix")
	outputDir := flag.String("outdir", ".", "Output directory")
	fileExt := flag.String("ext", "txt", "Output file extension")
	padWidth := flag.Int("pad", 3, "Zero padding width for file index")
	timestamp := flag.Bool("ts", false, "Add timestamp to filenames")
	dryRun := flag.Bool("dry", false, "Dry run mode (preview only)")
	quiet := flag.Bool("q", false, "Quiet mode (suppress logs)")

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
	if !*quiet {
		logInfo(fmt.Sprintf("📄 Input File: %s (%.2f MB)", *inputFile, float64(stat.Size())/(1024*1024)))
	}

	maxSizeBytes, err := parseSize(*sizePerFile)
	if err != nil {
		logWarn("Invalid size format: " + err.Error())
		maxSizeBytes = 0
	}

	var re *regexp.Regexp
	if *pattern != "" {
		re, err = regexp.Compile(*pattern)
		if err != nil {
			logError("Invalid regex pattern: " + err.Error())
			os.Exit(1)
		}
	}

	splitFile(file, *linesPerFile, maxSizeBytes, re, *outputDir, *outPrefix, *fileExt, *padWidth, *timestamp, *dryRun, *quiet)
}

func splitFile(file *os.File, maxLines int, maxSizeBytes int64, pattern *regexp.Regexp, outputDir, prefix, ext string, padWidth int, useTS, dryRun, quiet bool) {
	reader := bufio.NewReaderSize(file, bufSize)
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
			if !quiet {
				logInfo("[DryRun] Would create: " + filename)
			}
			return nil
		}
		f, err := os.Create(filename)
		if err != nil {
			return err
		}
		out = f
		writer = bufio.NewWriterSize(out, bufSize)
		if !quiet {
			logInfo("✂️  Creating: " + filename)
		}
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
		lineBytes, err := reader.ReadSlice('\n')
		if err == io.EOF {
			if len(lineBytes) > 0 {
				if dryRun == false {
					writer.Write(lineBytes)
				}
			}
			break
		}
		if err != nil {
			if errors.Is(err, bufio.ErrBufferFull) {
				if dryRun == false {
					writer.Write(lineBytes)
				}
				continue
			}
			logError("Error reading line: " + err.Error())
			break
		}

		if (maxLines > 0 && lineCount >= maxLines) ||
			(maxSizeBytes > 0 && written+int64(len(lineBytes)) > maxSizeBytes) ||
			(pattern != nil && pattern.Match(lineBytes)) {
			err := createNewPart()
			if err != nil {
				logError("Failed to create new part: " + err.Error())
				break
			}
		}

		if !dryRun {
			writer.Write(lineBytes)
		}
		lineCount++
		written += int64(len(lineBytes))
	}

	if !dryRun && writer != nil {
		writer.Flush()
		if out != nil {
			out.Close()
		}
	}

	if !quiet {
		logSuccess("🎉 Done! All parts created.")
	}
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
