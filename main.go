package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
)

const projectURL = "https://github.com/libai07/routecheck"
const ipInfoTimeout = 3 * time.Second
const traceTimeout = 20 * time.Second
const ordinaryRetryCount = 3

type IPInfo struct {
	City    string `json:"city"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

func main() {
	targetsFile := flag.String("targets", "", "自定义目标 JSON 文件路径")
	flag.Parse()

	targets, source, err := resolveTargets(*targetsFile)
	if err != nil {
		log.Fatal(err)
	}
	if len(targets) == 0 {
		log.Fatal("没有可测试的目标")
	}

	green := color.New(color.FgHiGreen).SprintFunc()
	cyan := color.New(color.FgHiCyan).SprintFunc()
	yellow := color.New(color.FgHiYellow).Add(color.Bold).SprintFunc()

	log.Println("正在测试三网回程路由...")

	info := fetchIPInfo()
	if info.Country != "" || info.City != "" || info.Org != "" {
		fmt.Println(green("国家: ") + cyan(info.Country) + green(" 城市: ") + cyan(info.City) + green(" 服务商: ") + cyan(info.Org))
	}
	fmt.Println(green("项目地址:"), yellow(projectURL))
	fmt.Println(green("测试来源:"), yellow(source))

	results := make([]Result, len(targets))
	ch := make(chan Result, len(targets))
	timeout := time.After(traceTimeout)

	for i := range targets {
		go trace(ch, i, targets[i])
	}

	completed := 0
	timedOut := false
	for completed < len(targets) && !timedOut {
		select {
		case r := <-ch:
			results[r.i] = r
			completed++
		case <-timeout:
			timedOut = true
		}
	}

	if timedOut {
		timeoutLabel := color.New(color.FgRed).Add(color.Bold).Sprint("超时")
		for i := range results {
			if results[i].s == "" {
				results[i] = Result{i: i, s: formatResult(targets[i], timeoutLabel)}
			}
		}
	}

	retryOrdinaryResults(results, targets)

	for _, result := range results {
		fmt.Println(result.s)
	}

	log.Println(green("测试完成!"))
}

func fetchIPInfo() IPInfo {
	client := &http.Client{Timeout: ipInfoTimeout}
	rsp, err := client.Get("http://ipinfo.io")
	if err != nil {
		return IPInfo{}
	}
	defer rsp.Body.Close()

	info := IPInfo{}
	if err := json.NewDecoder(rsp.Body).Decode(&info); err != nil {
		return IPInfo{}
	}
	return info
}

func resolveTargets(targetsFile string) ([]Target, string, error) {
	if targetsFile != "" {
		targets, err := loadTargetsFromFile(targetsFile)
		if err != nil {
			return nil, "", err
		}
		return targets, targetsFile, nil
	}

	return defaultTargets, "默认目标", nil
}

func loadTargetsFromFile(path string) ([]Target, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var targets []Target
	if err := json.Unmarshal(data, &targets); err != nil {
		var wrapped struct {
			Targets []Target `json:"targets"`
		}
		if err2 := json.Unmarshal(data, &wrapped); err2 != nil {
			return nil, fmt.Errorf("解析目标文件失败: %w", err)
		}
		targets = wrapped.Targets
	}

	valid := make([]Target, 0, len(targets))
	for _, target := range targets {
		if target.Name == "" || target.IP == "" {
			continue
		}
		if net.ParseIP(target.IP) == nil {
			return nil, fmt.Errorf("目标 %s 的 IP 格式无效: %s", target.Name, target.IP)
		}
		valid = append(valid, target)
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("目标文件没有有效目标: %s", path)
	}
	return valid, nil
}
