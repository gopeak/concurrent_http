package main

import (
	"bufio"
	"fmt"
	"github.com/valyala/fasthttp"
	"io"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var queryArr = []string{}

var Errorhttp []string

type SocketResponse struct {
	data     string
	response string
	err      error
}


func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println(os.Args)
	if len(os.Args) < 5 {
		fmt.Fprintf(os.Stderr, "Usage: host port query conn_num  times  ")
		os.Exit(1)
	}

	start := time.Now().Unix()
	go endHookStat(start)

	//golog.Println( "urls:" ,urls  )
	conn_num, _ := strconv.ParseInt(os.Args[4], 10, 32)
	results := asyncHttpReq(conn_num)

	i := 0
	for _, result := range results {
		i++
		fmt.Printf("%d result length: %d\n", i, len(result.data))
	}

	fmt.Printf(" result num: %d\n", len(results))

	select {}
}

func endHookStat(start int64) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	s := <-c
	end := time.Now().Unix()
	els_time := end - start
	fmt.Println("Got signal:", s)
	fmt.Println("need time :", els_time)
	fmt.Println("error http :", Errorhttp)

	os.Exit(1)
}

func asyncHttpReq(conn_num int64) []*SocketResponse {

	//ch := make( chan *SocketResponse )
	responses := []*SocketResponse{}

	for n := 0; n < int(conn_num); n++ {
		go func(n int) {
			reqWithSocketHttp(n)
		}(n)
	}

	return responses
}

func PrintBody(resp *fasthttp.Response){
	//fmt.Println(resp.Body())
}

func reqWithFastHttp(conn_num int) {

	host := os.Args[1]
	port := os.Args[2]
	run_times, _ := strconv.ParseInt(os.Args[5], 10, 32)
	query := "/"
	if len(os.Args) >= 3 {
		query = string(os.Args[3])
	}
	req_data := fmt.Sprintf("%s?i=%d", query, conn_num)

	for n := 0; n < int(run_times); n++ {
		url :=  fmt.Sprintf("http://%s:%s%s", host, port,req_data)
		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req) // 用完需要释放资源
		req.Header.SetUserAgent("Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36")
		req.Header.SetMethod("GET")
		req.SetRequestURI(url)
		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源

		if err := fasthttp.Do(req, resp); err != nil {
			fmt.Println("请求失败:", err.Error())
			break
		}
		PrintBody(resp)
		if resp.StatusCode()!=200 {
			fmt.Println("请求失败:", resp.StatusCode())
		}
		fmt.Println(url, n, ":", resp.StatusCode())

		time.Sleep(100 * time.Millisecond)
	}

}

func reqWithSocketHttp(conn_num int) {

	host := os.Args[1]
	port := os.Args[2]
	run_times, _ := strconv.ParseInt(os.Args[5], 10, 32)
	query := "/"
	if len(os.Args) >= 3 {
		query = string(os.Args[3])
	}
	req_data := fmt.Sprintf("%s?i=%d", query, conn_num)

	fmt.Println(req_data, run_times)
	requestHeader := fmt.Sprintf("GET http://%s:%s%s HTTP/1.1\r\nAccept-Language: zh-Hans-CN,zh-Hans;q=0.8,en-US;q=0.5,en;q=0.3\r\nUser-Agent:.User-Agent: Mozilla/5.0 (Windows NT 6.1; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/69.0.3497.100 Safari/537.36\r\nAccept-Encoding: gzip, deflate\r\nHost: question.shopwalker.cn:80\r\nConnection: close\r\n\r\n", host, port, req_data)
	//fmt.Println(requestHeader)

	for n := 0; n < int(run_times); n++ {

		// 建立连接
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
		if err != nil {
			fmt.Println("net.Dial error:", err.Error())
			return
		}
		defer conn.Close()
		// 向服务端发送数据。n返回数据大小，err返回错误信息。
		_, err = conn.Write([]byte(requestHeader))
		if err != nil {
			fmt.Println("retry net.Dial error:", err.Error())
			return

		}
		//fmt.Println("write size:", n)
		r := bufio.NewReader(conn)
		var response string
		j := 0
		header_first := ""
		for {
			j++
			item, err := r.ReadString('\n')
			if err != nil  {
				fmt.Println("r.ReadString error:", err.Error())
				break
			}
			// 第一行是状态码
			if j == 1 {
				header_first = item
				if strings.Index(item, "200") == -1 {
					Errorhttp = append(Errorhttp, header_first)
				}
				conn.Close()
				break
			}
			//读取内容
			if err == io.EOF {
				break
			}
			//fmt.Println("item:", item)
			response = response + item
		}

		fmt.Println(req_data, n, " :", header_first)

		time.Sleep(100 * time.Millisecond)
	}


}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
