package main

import (
	"flag"
	"fmt"
	"github.com/rakyll/hey/requester"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/http"
	gourl "net/url"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"time"
)

var (
	testing_path    = "testing.yml"
	testing_qps     = 50.0
	testing_number  = 200
	testing_count   = 50
	testing_timeout = 10
	testing_time    = time.Duration(0)
	m               = flag.String("m", "GET", "")
	headers         = flag.String("h", "", "")
	body            = flag.String("d", "", "")
	bodyFile        = flag.String("D", "", "")
	accept          = flag.String("A", "", "")
	contentType     = flag.String("T", "text/html", "")
	authHeader      = flag.String("a", "", "")
	hostHeader      = flag.String("host", "", "")

	output = flag.String("o", "", "")

	c = flag.Int("c", 50, "")
	n = flag.Int("n", 200, "")
	q = flag.Float64("q", 0, "")
	t = flag.Int("t", 20, "")
	z = flag.Duration("z", 0, "")

	h2   = flag.Bool("h2", false, "")
	cpus = flag.Int("cpus", runtime.GOMAXPROCS(-1), "")

	disableCompression = flag.Bool("disable-compression", false, "")
	disableKeepAlives  = flag.Bool("disable-keepalive", false, "")
	disableRedirects   = flag.Bool("disable-redirects", false, "")
	proxyAddr          = flag.String("x", "", "")
)

type SceneConf struct {
	Name string
	Urls []UrlsConf
}
type Conf struct {
	Globalheaders []GlobalheadersConf
	Scene         []SceneConf
}

type GlobalheadersConf struct {
	Key   string
	Value string
}
type UrlsConf struct {
	Url     string
	Method  string
	Name    string
	Data    string
	Headers []GlobalheadersConf
}

func main() {
	app := cli.NewApp()
	app.Name = "压力测试"
	app.Version = "1.0.0"
	app.Usage = "接口压力测试"

	app.Action = func(c *cli.Context) error {
		header := make(http.Header)
		data, _ := ioutil.ReadFile(testing_path)
		//fmt.Println(string(data))
		t := Conf{}
		//把yaml形式的字符串解析成struct类型
		yaml.Unmarshal(data, &t)

		// 获取配置文件中设置的headers
		for headerIndex := 0; headerIndex < len(t.Globalheaders); headerIndex++ {
			_header := t.Globalheaders[headerIndex]
			header.Set(_header.Key, _header.Value)
		}

		// 开始批量压测
		for sceneIndex := 0; sceneIndex < len(t.Scene); sceneIndex++ {
			scene := t.Scene[sceneIndex]
			fmt.Println("开始压测场景：", scene.Name)
			for urlIndex := 0; urlIndex < len(scene.Urls); urlIndex++ {
				_url := scene.Urls[urlIndex]
				fmt.Println("压测地址：", _url.Name)
				url := _url.Url
				method := _url.Method
				req, err := http.NewRequest(method, url, nil)
				if err != nil {
					usageAndExit(err.Error())
				}

				// 处理POST数据
				var bodyAll []byte
				if _url.Data != "" {
					bodyAll = []byte(_url.Data)
				}

				var proxyURL *gourl.URL
				dur := testing_time
				// 设置场景里URL设置的headers
				for _urlHeadersIndex := 0; _urlHeadersIndex < len(_url.Headers); _urlHeadersIndex++ {
					_header := _url.Headers[_urlHeadersIndex]
					header.Set(_header.Key, _header.Value)
				}
				req.Header = header
				w := &requester.Work{
					Request:            req,
					RequestBody:        bodyAll,
					N:                  testing_number,
					C:                  testing_count,
					QPS:                testing_qps,
					Timeout:            testing_timeout,
					DisableCompression: *disableCompression,
					DisableKeepAlives:  *disableKeepAlives,
					DisableRedirects:   *disableRedirects,
					H2:                 *h2,
					ProxyAddr:          proxyURL,
					Output:             *output,
				}
				w.Init()
				c := make(chan os.Signal, 1)
				signal.Notify(c, os.Interrupt)
				go func() {
					<-c
					w.Stop()
				}()
				if dur > 0 {
					go func() {
						time.Sleep(dur)
						w.Stop()
					}()
				}
				w.Run()

			}
		}
		return nil
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "file, f",
			Usage:       "配置文件路径",
			Value:       "testing.yml",
			Destination: &testing_path,
		},
		cli.Float64Flag{
			Name:        "qps, q",
			Usage:       "每秒请求次数",
			Value:       50.0,
			Destination: &testing_qps,
		},
		cli.IntFlag{
			Name:        "number, n",
			Usage:       "请求总数",
			Value:       200,
			Destination: &testing_number,
		},

		cli.IntFlag{
			Name:        "count, c",
			Usage:       "并发数",
			Value:       200,
			Destination: &testing_count,
		},
		cli.DurationFlag{
			Name:        "time, t",
			Usage:       "持续时间，设置这个值时请求总数失效",
			Value:       0,
			Destination: &testing_time,
		},
		cli.IntFlag{
			Name:        "timeout",
			Usage:       "超时时间",
			Value:       10,
			Destination: &testing_timeout,
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}

	os.Exit(0)
}

func errAndExit(msg string) {
	fmt.Fprintf(os.Stderr, msg)
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func parseInputWithRegexp(input, regx string) ([]string, error) {
	re := regexp.MustCompile(regx)
	matches := re.FindStringSubmatch(input)
	if len(matches) < 1 {
		return nil, fmt.Errorf("could not parse the provided input; input = %v", input)
	}
	return matches, nil
}

type headerSlice []string

func (h *headerSlice) String() string {
	return fmt.Sprintf("%s", *h)
}

func (h *headerSlice) Set(value string) error {
	*h = append(*h, value)
	return nil
}
