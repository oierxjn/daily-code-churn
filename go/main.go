package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

type DayStat struct {
	Date    time.Time
	Added   int
	Removed int
}

func main() {
	var (
		days    = flag.Int("days", 30, "How many days to include")
		branch  = flag.String("branch", "", "Branch/ref to analyze (optional)")
		outPath = flag.String("out", "daily-churn.svg", "Output SVG path")
		width   = flag.Int("width", 1000, "SVG width")
		height  = flag.Int("height", 320, "SVG height")
	)
	flag.Parse()

	stats, err := collectDaily(*days, *branch)
	if err != nil {
		fatal(err)
	}

	svg := renderSVG(stats, *width, *height)

	if err := os.MkdirAll(dirOf(*outPath), 0o755); err != nil && dirOf(*outPath) != "." {
		fatal(err)
	}
	if err := os.WriteFile(*outPath, []byte(svg), 0o644); err != nil {
		fatal(err)
	}
	fmt.Println("Wrote", *outPath)
}

func dirOf(p string) string {
	i := strings.LastIndex(p, "/")
	if i < 0 {
		return "."
	}
	if i == 0 {
		return "/"
	}
	return p[:i]
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

// 收集指定天数内的每日代码变更统计
//
// 返回指定天数内的每日代码变更统计数组
func collectDaily(days int, branch string) ([]DayStat, error) {
	// git log --since=XX.days --date=short --pretty=format:@@@%ad --numstat [branch]
	args := []string{"log", fmt.Sprintf("--since=%d.days", days), "--date=short", "--pretty=format:@@@%ad", "--numstat"}
	if branch != "" {
		args = append(args, branch)
	}

	cmd := exec.Command("git", args...)
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		// include stderr if possible
		if ee, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git log failed: %v: %s", err, string(ee.Stderr))
		}
		return nil, fmt.Errorf("git log failed: %v", err)
	}

	added := map[string]int{}
	removed := map[string]int{}
	var curDate string

	// 创建扫描器
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "@@@") {
			curDate = strings.TrimSpace(strings.TrimPrefix(line, "@@@"))
			continue
		}
		if strings.TrimSpace(line) == "" || curDate == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}
		a, d := parts[0], parts[1]
		if a == "-" || d == "-" {
			continue // binary
		}
		ai, err1 := strconv.Atoi(a)
		di, err2 := strconv.Atoi(d)
		if err1 != nil || err2 != nil {
			continue
		}
		added[curDate] += ai
		removed[curDate] += di
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	// Build continuous date range ending today (UTC) so missing days show as 0
	today := time.Now().UTC().Truncate(24 * time.Hour)
	start := today.AddDate(0, 0, -(days - 1))

	var res []DayStat
	for d := start; !d.After(today); d = d.AddDate(0, 0, 1) {
		key := d.Format(time.DateOnly)
		res = append(res, DayStat{
			Date:    d,
			Added:   added[key],
			Removed: removed[key],
		})
	}

	// ensure stable ordering
	sort.Slice(res, func(i, j int) bool { return res[i].Date.Before(res[j].Date) })
	return res, nil
}

func renderSVG(days []DayStat, width, height int) string {
	// Layout
	const (
		padL = 64
		padR = 12
		padT = 12
		padB = 28
	)
	innerW := width - padL - padR
	innerH := height - padT - padB
	baselineY := padT + innerH/2

	// Find max to scale bars
	maxV := 0
	for _, d := range days {
		if d.Added > maxV {
			maxV = d.Added
		}
		if d.Removed > maxV {
			maxV = d.Removed
		}
	}
	if maxV == 0 {
		maxV = 1
	}
	maxBarH := innerH/2 - 8
	scale := float64(maxBarH) / float64(maxV)

	n := len(days)
	if n == 0 {
		return emptySVG(width, height)
	}

	groupW := float64(innerW) / float64(n)
	barW := int(groupW * 0.68)
	if barW < 2 {
		barW = 2
	}
	if barW > 18 {
		barW = 18
	}

	var svgBuilder strings.Builder
	svgBuilder.Grow(32 * 1024)
	svgBuilder.WriteString(fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height))
	svgBuilder.WriteString(`<rect width="100%" height="100%" fill="white"/>`)

	// grid lines (optional, light)
	for i := 1; i <= 4; i++ {
		yUp := baselineY - int(float64(maxBarH)*float64(i)/4.0)
		yDn := baselineY + int(float64(maxBarH)*float64(i)/4.0)
		svgBuilder.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#eee"/>`, padL, yUp, width-padR, yUp))
		svgBuilder.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#eee"/>`, padL, yDn, width-padR, yDn))
		tick := maxV * i / 4
		svgBuilder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="10" text-anchor="end" dominant-baseline="middle" fill="#666">+%s</text>`, padL-6, yUp, human(tick)))
		svgBuilder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="10" text-anchor="end" dominant-baseline="middle" fill="#666">-%s</text>`, padL-6, yDn, human(tick)))
	}

	// baseline
	svgBuilder.WriteString(fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#333"/>`, padL, baselineY, width-padR, baselineY))
	svgBuilder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="10" text-anchor="end" dominant-baseline="middle" fill="#666">0</text>`, padL-6, baselineY))

	// bars
	for i, d := range days {
		xCenter := float64(padL) + (float64(i)+0.5)*groupW
		x := int(xCenter) - barW/2

		addH := int(float64(d.Added) * scale)
		remH := int(float64(d.Removed) * scale)

		// Added (upwards)
		if addH > 0 {
			y := baselineY - addH
			svgBuilder.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#6cc04a"/>`, x, y, barW, addH))
		}
		// Removed (downwards)
		if remH > 0 {
			y := baselineY
			svgBuilder.WriteString(fmt.Sprintf(`<rect x="%d" y="%d" width="%d" height="%d" fill="#e5533d"/>`, x, y, barW, remH))
		}
	}

	// X axis labels (sparse)
	labelEvery := 1
	if n > 45 {
		labelEvery = 7
	} else if n > 20 {
		labelEvery = 3
	}
	for i, d := range days {
		if i%labelEvery != 0 && i != n-1 {
			continue
		}
		x := float64(padL) + (float64(i)+0.5)*groupW
		lbl := d.Date.Format("01-02")
		svgBuilder.WriteString(fmt.Sprintf(`<text x="%.1f" y="%d" font-size="10" text-anchor="middle" fill="#666">%s</text>`, x, height-10, lbl))
	}

	// Title
	sumA, sumR := 0, 0
	for _, d := range days {
		sumA += d.Added
		sumR += d.Removed
	}
	svgBuilder.WriteString(fmt.Sprintf(`<text x="%d" y="%d" font-size="12" fill="#111">Daily code churn (%d days): +%s  -%s</text>`,
		padL, 18, n, human(sumA), human(sumR)))

	svgBuilder.WriteString(`</svg>`)
	return svgBuilder.String()
}

// 将数字格式化为人类易读的字符串
//
// 例如：240130 -> 240.1k
func human(n int) string {
	// simple compact format: 240130 -> 240.1k
	if n < 1000 {
		return strconv.Itoa(n)
	}
	if n < 1000000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000.0)
	}
	return fmt.Sprintf("%.2fM", float64(n)/1000000.0)
}

func emptySVG(width, height int) string {
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d"><rect width="100%%" height="100%%" fill="white"/><text x="10" y="20" font-size="12">No data</text></svg>`, width, height)
}
