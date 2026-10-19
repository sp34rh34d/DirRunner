package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"text/tabwriter"
	"time"
)

var Verbose bool
var StreamResults bool

// Version is set from main at startup and shown in the banner.
var Version = "2.0.0"

// termMu serializes writes to the terminal so the live progress line (stderr)
// never interleaves with result lines (stdout) or log messages.
var termMu sync.Mutex

// progressTTY reports whether stderr is an interactive terminal. The live
// progress bar (with carriage returns) is only drawn when it is; otherwise the
// output is a plain, pipe-friendly stream with no control characters.
var progressTTY = isTerminal(os.Stderr)

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// ANSI color codes and whether color is allowed on each stream. Color is
// disabled when the stream is not a terminal, or when NO_COLOR / TERM=dumb is
// set, so piped output and files stay clean.
const (
	ansiReset   = "\033[0m"
	ansiBold    = "\033[1m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiBlue    = "\033[34m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
	ansiGray    = "\033[90m"
	// Neon (bright) variants drive the cyberpunk theme.
	ansiNeonPink   = "\033[95m"
	ansiNeonCyan   = "\033[96m"
	ansiNeonGreen  = "\033[92m"
	ansiNeonYellow = "\033[93m"
	ansiNeonRed    = "\033[91m"
	ansiNeonBlue   = "\033[94m"
)

var (
	colorStderr = colorAllowed(os.Stderr)
	colorStdout = colorAllowed(os.Stdout)
)

func colorAllowed(f *os.File) bool {
	if os.Getenv("NO_COLOR") != "" || os.Getenv("TERM") == "dumb" {
		return false
	}
	return isTerminal(f)
}

// paint wraps s in an ANSI color when enabled; otherwise it returns s unchanged.
func paint(enabled bool, code, s string) string {
	if !enabled || code == "" {
		return s
	}
	return code + s + ansiReset
}

// statusColor maps an HTTP status code to a neon color by class.
func statusColor(code int) string {
	switch {
	case code >= 200 && code < 300:
		return ansiNeonGreen
	case code >= 300 && code < 400:
		return ansiNeonCyan
	case code >= 400 && code < 500:
		return ansiNeonYellow
	case code >= 500:
		return ansiNeonRed
	default:
		return ansiGray
	}
}

// typeColor gives each result type a stable neon accent color.
func typeColor(t string) string {
	switch t {
	case "directory", "file", "url":
		return ansiNeonBlue
	case "dns", "technology", "fingerprint", "header":
		return ansiNeonCyan
	case "traversal", "sqli", "bypass", "dbms":
		return ansiNeonPink
	case "email", "match", "js", "img", "clone":
		return ansiNeonGreen
	default:
		return ansiBold + ansiNeonCyan
	}
}

// clearProgressLine erases the current in-place progress line so the next
// message prints cleanly. No-op when stderr is not a terminal. Callers must
// hold termMu.
func clearProgressLine() {
	if progressTTY {
		fmt.Fprint(os.Stderr, "\r\033[K")
	}
}

func emit(prefixColor, prefix, msgColor, format string, args ...any) {
	termMu.Lock()
	defer termMu.Unlock()
	clearProgressLine()
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s%s\n", paint(colorStderr, prefixColor, prefix), paint(colorStderr, msgColor, msg))
}

type Result struct {
	Type     string `json:"type"`
	Target   string `json:"target"`
	Method   string `json:"method,omitempty"`
	URL      string `json:"url,omitempty"`
	Path     string `json:"path,omitempty"`
	Host     string `json:"host,omitempty"`
	Status   int    `json:"status,omitempty"`
	Size     int64  `json:"size,omitempty"`
	Location string `json:"location,omitempty"`
	IP       string `json:"ip,omitempty"`
	Value    string `json:"value,omitempty"`
}

func Banner(w io.Writer) {
	const width = 66
	top := "╔" + strings.Repeat("═", width-2) + "╗"
	bottom := "╚" + strings.Repeat("═", width-2) + "╝"
	art := []string{
		"   ___  _     ___",
		"  / _ \\(_)___/ _ \\__ _____  ___  ___ ____",
		" / // / / __/ , _/ // / _ \\/ _ \\/ -_) __/",
		"/____/_/_/ /_/|_|\\_,_/_//_/_//_/\\__/_/",
	}
	fmt.Fprintln(w, paint(colorStderr, ansiNeonPink, top))
	for _, line := range art {
		fmt.Fprintln(w, paint(colorStderr, ansiNeonPink+ansiBold, line))
	}
	fmt.Fprintln(w, paint(colorStderr, ansiNeonCyan+ansiBold, "  ░▒▓ D I R R U N N E R ▓▒░  ")+
		paint(colorStderr, ansiNeonGreen, "v"+Version))
	fmt.Fprintln(w, paint(colorStderr, ansiGray, "  sp34rh34d x @spearh34d"))
	fmt.Fprintln(w, paint(colorStderr, ansiNeonPink, bottom))
}

func OptionsMenu(w io.Writer, title string, rows [][2]string) error {
	Banner(w)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintf(tw, "%s\t%s\n", paint(colorStderr, ansiNeonPink+ansiBold, "MODULE"), paint(colorStderr, ansiNeonCyan+ansiBold, title))
	fmt.Fprintf(tw, "%s\t%s\n", paint(colorStderr, ansiGray, "OPTION"), paint(colorStderr, ansiGray, "VALUE"))
	for _, row := range rows {
		if row[1] == "" {
			continue
		}
		fmt.Fprintf(tw, "%s\t%s\n", paint(colorStderr, ansiNeonCyan, row[0]), row[1])
	}
	fmt.Fprintln(tw, strings.Repeat("-", 66)+"\t")
	return tw.Flush()
}

func Info(format string, args ...any) {
	emit(ansiNeonCyan, "[◇] ", "", format, args...)
}

func Warn(format string, args ...any) {
	emit(ansiNeonPink, "[▲] ", ansiNeonPink, format, args...)
}

// Success prints a positive, neon-green message.
func Success(format string, args ...any) {
	emit(ansiNeonGreen, "[✓] ", ansiNeonGreen, format, args...)
}

func Debug(format string, args ...any) {
	if Verbose {
		emit(ansiGray, "[»] ", ansiGray, format, args...)
	}
}

// progressRefresh throttles how often the live progress line is redrawn so a
// fast scan does not flood the terminal with writes.
const progressRefresh = 100 * time.Millisecond

type ProgressTracker struct {
	label     string
	total     int
	current   int
	lastPrint time.Time
	mu        sync.Mutex
}

func NewProgress(label string, total int) *ProgressTracker {
	return &ProgressTracker{label: label, total: total}
}

func (p *ProgressTracker) Advance() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current++
	// The in-place bar only makes sense on an interactive terminal.
	if progressTTY && time.Since(p.lastPrint) >= progressRefresh {
		p.print()
		p.lastPrint = time.Now()
	}
}

func (p *ProgressTracker) SetTotal(total int) {
	if p == nil || total <= 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total = total
}

// AddTotal grows the expected total, used when recursive discovery queues more
// work while the scan is already running.
func (p *ProgressTracker) AddTotal(delta int) {
	if p == nil || delta <= 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.total += delta
}

func (p *ProgressTracker) Finish() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	termMu.Lock()
	defer termMu.Unlock()
	clearProgressLine()
	fmt.Fprintf(os.Stderr, "%s %s\n", paint(colorStderr, ansiNeonCyan, "[◇]"), p.summary())
}

// print draws the in-place progress line. Callers hold p.mu; it is only
// reached when progressTTY is true.
func (p *ProgressTracker) print() {
	termMu.Lock()
	defer termMu.Unlock()
	fmt.Fprintf(os.Stderr, "\r\033[K%s %s", paint(colorStderr, ansiNeonCyan, "[◇]"), p.summary())
}

// progressBarWidth is the number of cells in the drawn progress bar.
const progressBarWidth = 24

func (p *ProgressTracker) summary() string {
	if p.total > 0 {
		percent := p.current * 100 / p.total
		filled := percent * progressBarWidth / 100
		if filled > progressBarWidth {
			filled = progressBarWidth
		}
		bar := paint(colorStderr, ansiNeonPink, strings.Repeat("█", filled)) +
			paint(colorStderr, ansiGray, strings.Repeat("░", progressBarWidth-filled))
		return fmt.Sprintf("%s ⟦%s⟧ %s %s",
			paint(colorStderr, ansiNeonCyan, p.label),
			bar,
			paint(colorStderr, ansiNeonGreen, fmt.Sprintf("%3d%%", percent)),
			paint(colorStderr, ansiGray, fmt.Sprintf("(%d/%d)", p.current, p.total)))
	}
	return fmt.Sprintf("%s: %d processed", p.label, p.current)
}

// writeResultLine prints a streamed result to stdout, first erasing any live
// progress line on the terminal so the two never collide.
func writeResultLine(line string) {
	termMu.Lock()
	defer termMu.Unlock()
	clearProgressLine()
	fmt.Fprintln(os.Stdout, line)
}

func PrintLiveResult(result Result) {
	if !StreamResults {
		return
	}
	target := displayTarget(result)
	if Verbose && result.Type == "dns" {
		details := []string{"host=" + target}
		if result.IP != "" {
			details = append(details, "ip="+result.IP)
		}
		Debug("found %s", strings.Join(details, " "))
		return
	}
	if result.Type == "dns" {
		typeField := paint(colorStdout, typeColor(result.Type), fmt.Sprintf("%-10s", result.Type))
		writeResultLine(fmt.Sprintf("%s %-32s %s", typeField, target, result.IP))
		return
	}
	// Extraction findings (crawl --emails / --regx) carry no meaningful status or
	// size; show the captured value and the page it was found on instead.
	if result.Value != "" && (result.Type == "email" || result.Type == "match") {
		typeField := paint(colorStdout, typeColor(result.Type), fmt.Sprintf("%-10s", result.Type))
		valueField := paint(colorStdout, ansiNeonGreen, fmt.Sprintf("%-40s", result.Value))
		writeResultLine(fmt.Sprintf("%s %s %s", typeField, valueField, target))
		return
	}
	typeField := paint(colorStdout, typeColor(result.Type), fmt.Sprintf("%-10s", result.Type))
	statusField := paint(colorStdout, statusColor(result.Status), fmt.Sprintf("%-7d", result.Status))
	writeResultLine(fmt.Sprintf("%s %s %-8d %s", typeField, statusField, result.Size, target))
	if Verbose {
		details := []string{
			"type=" + result.Type,
			"status=" + fmt.Sprint(result.Status),
			"size=" + fmt.Sprint(result.Size),
			"target=" + target,
		}
		if result.URL != "" && result.URL != target {
			details = append(details, "url="+result.URL)
		}
		if result.Method != "" {
			details = append(details, "method="+result.Method)
		}
		if result.Location != "" {
			details = append(details, "location="+result.Location)
		}
		if result.Host != "" {
			details = append(details, "host="+result.Host)
		}
		if result.IP != "" {
			details = append(details, "ip="+result.IP)
		}
		if result.Value != "" {
			details = append(details, "value="+result.Value)
		}
		Debug("found %s", strings.Join(details, " "))
	}
}

func PrintResults(w io.Writer, results []Result, jsonOut bool) error {
	if jsonOut {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}
	if isFingerprintResults(results) {
		return PrintFingerprintResults(w, results)
	}
	if isDNSResults(results) {
		return PrintDNSResults(w, results)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TYPE\tSTATUS\tSIZE\tTARGET\tLOCATION")
	for _, r := range results {
		target := displayTarget(r)
		if r.IP != "" {
			target = fmt.Sprintf("%s -> %s", target, r.IP)
		}
		if r.Value != "" && r.Type != "fuzz" {
			target = fmt.Sprintf("%s (%s)", target, r.Value)
		}
		status := ""
		if r.Status > 0 {
			status = fmt.Sprint(r.Status)
		}
		size := ""
		if r.Size >= 0 {
			size = fmt.Sprint(r.Size)
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", r.Type, status, size, target, r.Location)
	}
	return tw.Flush()
}

func displayTarget(result Result) string {
	if (result.Type == "directory" || result.Type == "file") && result.Path != "" {
		return result.Path
	}
	if result.Type == "fuzz" && result.Value != "" && result.URL == result.Target {
		return result.Value
	}
	if result.URL != "" {
		return result.URL
	}
	if result.Host != "" {
		return result.Host
	}
	return result.Target
}

func isDNSResults(results []Result) bool {
	return len(results) > 0 && results[0].Type == "dns"
}

func PrintDNSResults(w io.Writer, results []Result) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "TYPE\tHOST\tIP")
	for _, result := range results {
		host := result.Host
		if host == "" {
			host = result.Target
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\n", result.Type, host, result.IP)
	}
	return tw.Flush()
}

func isFingerprintResults(results []Result) bool {
	for _, result := range results {
		if result.Type == "fingerprint" || result.Type == "technology" || result.Type == "header" {
			return true
		}
	}
	return false
}

func PrintFingerprintResults(w io.Writer, results []Result) error {
	var summary *Result
	var technologies []Result
	var headers []Result
	for i := range results {
		switch results[i].Type {
		case "fingerprint":
			summary = &results[i]
		case "technology":
			technologies = append(technologies, results[i])
		case "header":
			headers = append(headers, results[i])
		}
	}
	sort.Slice(technologies, func(i, j int) bool {
		return technologies[i].Value < technologies[j].Value
	})

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "SUMMARY\t")
	if summary != nil {
		fmt.Fprintf(tw, "Target\t%s\n", summary.URL)
		fmt.Fprintf(tw, "Status\t%d\n", summary.Status)
		fmt.Fprintf(tw, "Size\t%d\n", summary.Size)
	}

	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "TECHNOLOGIES\t")
	if len(technologies) == 0 {
		fmt.Fprintln(tw, "None detected\t")
	} else {
		fmt.Fprintln(tw, "Name\t")
		for _, tech := range technologies {
			fmt.Fprintf(tw, "%s\t\n", tech.Value)
		}
	}

	fmt.Fprintln(tw)
	fmt.Fprintln(tw, "HEADERS\t")
	if len(headers) == 0 {
		fmt.Fprintln(tw, "None\t")
	} else {
		fmt.Fprintln(tw, "Header\tValue")
		for _, header := range headers {
			name, value := splitHeaderValue(header.Value)
			fmt.Fprintf(tw, "%s\t%s\n", name, value)
		}
	}
	return tw.Flush()
}

func splitHeaderValue(value string) (string, string) {
	name, headerValue, ok := strings.Cut(value, ":")
	if !ok {
		return value, ""
	}
	return strings.TrimSpace(name), strings.TrimSpace(headerValue)
}

func WriteResults(path string, results []Result, jsonOut bool) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return PrintResults(f, results, jsonOut)
}
