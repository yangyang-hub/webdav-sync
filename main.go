package main

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gocolly/colly/v2"
)

const LocalPath string = "/data"
const TotalRetry int = 5

var BaseUrl string
var TaskFlag bool = false

func main() {
	BaseUrl = os.Getenv("WEBDAV_URL")
	if BaseUrl == "" {
		BaseUrl = "http://127.0.0.1:8888"
	}
	syncCorn := os.Getenv("SYNC_CORN")
	if syncCorn == "" {
		syncCorn = "*/30 * * * *"
	}
	fmt.Println("webdav url:", BaseUrl)
	webdavSync()
	timezone, _ := time.LoadLocation("Asia/Shanghai")
	scheduler := gocron.NewScheduler(timezone)
	_, err := scheduler.Cron(syncCorn).Do(webdavSync)
	if err != nil {
		fmt.Println("Error scheduling task:", err)
		return
	}
	scheduler.StartBlocking()
}

func webdavSync() {
	if TaskFlag {
		return
	}
	TaskFlag = true
	rmrfCommand := []string{}
	mkdirCommand := []string{}
	wgetCommand := [][]string{}
	files := getWebDavFiles(BaseUrl)
	localFiles := localPath()
	for key, linkMap := range localFiles {
		webdav := files[key]
		if len(webdav) == 0 {
			path := linkMap["path"]
			rmrfCommand = append(rmrfCommand, append_string([]string{LocalPath, path}))
		}
	}
	for _, value := range rmrfCommand {
		fmt.Println("delete file&path:", value)
		execCommand(0, "rm", "-rf", value)
	}

	for key, linkMap := range files {
		local := localFiles[key]
		if len(local) == 0 {
			path := linkMap["path"]
			file := linkMap["file"]
			if file == "Y" {
				link := linkMap["link"]
				wgetCommand = append(wgetCommand, []string{append_string([]string{LocalPath, path}), link})
			} else {
				mkdirCommand = append(mkdirCommand, append_string([]string{LocalPath, path}))
			}
		}
	}
	for _, value := range mkdirCommand {
		fmt.Println("create dir:", value)
		execCommand(0, "mkdir", "-p", value)
	}
	for _, value := range wgetCommand {
		fmt.Println("download file:", value[0], value[1])
		execCommand(0, "wget", "-O", value[0], value[1])
		time.Sleep(10 * time.Second)
	}
	TaskFlag = false
}

func getWebDavFiles(baseUrl string) map[string]map[string]string {
	files := map[string]map[string]string{}
	c := colly.NewCollector()
	c.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36 Edge/16.16299"
	c.OnHTML("table > tbody > tr", func(e *colly.HTMLElement) {
		link := ""
		text := ""
		path := ""
		dir := false
		e.ForEach("td", func(index int, col *colly.HTMLElement) {
			if index == 0 {
				h := col.ChildAttr("a", "href")
				text = col.ChildText("a")
				if len(h) > 1 && text != "Parent Directory" {
					p, _ := url.QueryUnescape(h)
					path = p
					link = append_string([]string{BaseUrl, p})
				}
			}
			if index == 2 {
				text := col.Text
				if text == "[DIR]    " {
					path = substr(path, 0, len([]rune(path))-1)
					dir = true
				} else {
					dir = false
				}
			}
		})
		if len(link) > 1 {
			linkMap := map[string]string{}
			linkMap["mame"] = text
			linkMap["path"] = path
			linkMap["link"] = link
			if dir {
				linkMap["file"] = "N"
				child := getWebDavFiles(link)
				for key, value := range child {
					files[key] = value
				}
			} else {
				linkMap["file"] = "Y"
			}
			files[path] = linkMap
		}
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Url visit wrong:", r.Request.URL, err)
	})
	c.Visit(baseUrl)
	c.Wait()
	return files
}

func localPath() map[string]map[string]string {
	files := map[string]map[string]string{}
	err := filepath.Walk(LocalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if err != nil {
			fmt.Println(err)
			return nil
		}
		path = substr(path, len([]rune(LocalPath)), len([]rune(path)))
		if len(path) > 1 {
			linkMap := map[string]string{}
			linkMap["mame"] = info.Name()
			linkMap["path"] = path
			if info.IsDir() {
				linkMap["file"] = "N"
			} else {
				linkMap["file"] = "Y"
			}
			files[path] = linkMap
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return files
}

func execCommand(retryCount int, name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
		if retryCount < TotalRetry {
			retryCount++
			time.Sleep(10 * time.Second)
			execCommand(retryCount, name, arg...)
		}
	}
	fmt.Println(string(output))
}

func append_string(sli []string) string {
	if len(sli) == 0 {
		return ""
	}
	var buffer bytes.Buffer
	for _, v := range sli {
		buffer.WriteString(v)
	}
	return buffer.String()
}

func substr(str string, start, length int) string {
	rs := []rune(str)
	rl := len(rs)
	end := 0

	if start < 0 {
		start = rl - 2 + start
	}
	end = start + length

	if start > end {
		start, end = end, start
	}

	if start < 0 {
		start = 0
	}
	if start > rl {
		start = rl
	}
	if end < 0 {
		end = 0
	}
	if end > rl {
		end = rl
	}

	return string(rs[start:end])
}
