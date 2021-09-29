package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
)

var update bool
var keywords string
var environment string = "prod"

func main() {
	//读取命令行输入
	parseInput()
	log.WithFields(log.Fields{"u": update, "kw": keywords}).Debug("parse args finished")

	//读取indexmap
	readIndexMap()

	//抓包并解析，然后更新索引
	if update {
		fmt.Println("start to update index")
		updateIndexes()
		fmt.Println("update finished")
	}
	fmt.Println("start to search")
	r := doSearch()
	if len(r) == 0 {
		log.WithFields(log.Fields{
			"indexmap": indexmap,
			"keywords": keywords,
		}).Info("search no match")
		fmt.Println("search no match,bye")
		return
	}
	//展示给用户
	showToUser(r)
	// for _, v := range s {
	// 	fmt.Print(v)
	// }
}

//设置日志配置
func init() {
	if environment != "debug" {
		log.SetFormatter(&log.TextFormatter{})
		file, err := os.OpenFile("xkcd_searchtool.log", os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			log.Warn("open log file error,use stdout")
			log.SetOutput(os.Stdout)
		} else {
			log.SetOutput(file)
		}
	} else {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stdout)
	}

	log.SetLevel(log.DebugLevel)
}

//读取命令行输入
func parseInput() {
	flag.BoolVar(&update, "u", true, "decide whether update the index")
	flag.StringVar(&keywords, "kw", "", "search key words")
	flag.Parse()
	if keywords == "" {
		log.Fatal("need keywords")
	}
}

const (
	INDEXNUM_FNAME  = "xkcd_indexnum.txt"
	REALINDEX_FNAME = "xkcd_realindex.txt"
)

const (
	URL_PREFIX = "https://xkcd.com/"
	URL_SUFFIX = "/info.0.json"
)

type XKCDResult struct {
	// month      string
	Num int
	// link       string
	// year       string
	// news       string
	// safeTitle  string `json:"safe_title"`
	Transcript string
	// alt        string
	// imt        string
	// title      string
	// day        string
}

var indexmap map[string][]int

func readIndexMap() {
	//读取索引文件
	indexmap = map[string][]int{}
	_, err := os.Stat(REALINDEX_FNAME)
	if err != nil {
		if os.IsNotExist(err) {
			//文件不存在，默认为初次运行
			log.Warn("realindex file is not exist,start with empty map")
		} else {
			//其他错误，终止程序运行
			log.WithError(err).WithField("REALINDEX_FNAME", REALINDEX_FNAME).Fatal("os.Stat err")
		}
	} else {
		file, err := os.Open(REALINDEX_FNAME)
		if err != nil {
			//其他错误，终止程序运行
			log.WithError(err).WithField("REALINDEX_FNAME", REALINDEX_FNAME).Fatal("os.Open err")
		}
		//读取json文件到map中
		json.NewDecoder(file).Decode(&indexmap)
		// for k, v := range indexmap {
		// 	fmt.Println(k, v)
		// }
		log.Debug("process indexmap finished")
	}
}

//抓包并解析，然后更新索引
func updateIndexes() {
	//读取编号文件,确定索引号
	i := 1
	var si string
	_, err := os.Stat(INDEXNUM_FNAME)
	if err != nil {
		if os.IsNotExist(err) {
			//文件不存在，默认为初次运行
			log.Warn("indexnum file is not exist,start with 1")
		} else {
			//其他错误，终止程序运行
			log.WithError(err).WithField("INDEXNUM_FNAME", INDEXNUM_FNAME).Fatal("os.Stat err")
		}
	} else {
		datas, err := ioutil.ReadFile(INDEXNUM_FNAME)
		if err != nil {
			//其他错误，终止程序运行
			log.WithError(err).WithField("INDEXNUM_FNAME", INDEXNUM_FNAME).Fatal("ioutil.ReadFile err")
		}
		si = string(datas)
		i, err = strconv.Atoi(si)
		if err != nil {
			log.WithError(err).WithField("indexnum", si).Fatal("strconv.Atoi err")
		}
	}
	log.WithField("i", i).Debug("process i finished")

	//抓包，解析
	for {
		//404 特殊处理
		if i == 404 {
			i++
			continue
		}
		//发送请求
		url := URL_PREFIX + strconv.Itoa(i) + URL_SUFFIX
		rp, err := http.Get(url)
		if err != nil {
			log.WithError(err).WithField("url", url).Fatal("http.Get err")
		}

		if rp.StatusCode == http.StatusNotFound {
			log.WithField("url", url).Debug("finish updated")
			break
		}

		var xkcdResult *XKCDResult
		err = json.NewDecoder(rp.Body).Decode(&xkcdResult)
		if err != nil {
			rp.Body.Close()
			log.WithError(err).WithField("url", url).Fatal("http.Get err")
		}
		log.WithFields(log.Fields{
			"num":        xkcdResult.Num,
			"transcript": xkcdResult.Transcript,
			"url":        url,
		}).Debug("http finished")

		//如果得到的结果不是正规格式，略过
		if !strings.HasPrefix(xkcdResult.Transcript, "[[") {
			log.WithFields(log.Fields{
				"num":        xkcdResult.Num,
				"transcript": xkcdResult.Transcript,
				"url":        url,
			}).Error("wrong transcript omit it")
			i++
			continue
		}

		//更新索引
		scan := bufio.NewScanner(strings.NewReader(xkcdResult.Transcript))
		scan.Split(scanScriptLines)

		for scan.Scan() {
			s := scan.Text()
			linecontext := log.WithField("scan_text", s)
			linecontext.Trace("scan Text finished")
			//按空格分隔
			words := strings.Split(s, " ")

			for _, word := range words {
				//如果不是单词,跳过不存
				if len(word) == 0 || !unicode.IsLetter(rune(word[0])) {
					continue
				}
				wordcontext := linecontext.WithField("word", word)
				wordcontext.Trace("scan word finished")
				//处理一下末尾的,和.
				sp := strings.TrimRight(word, ".,")
				wordcontext.WithField("trim_word", sp).Trace("trim finished")

				docs, ok := indexmap[sp]
				if !ok {
					b := make([]int, 0, xkcdResult.Num*2)
					b = append(b, xkcdResult.Num)
					indexmap[sp] = b
				} else {
					//判断漫画号是否已存在，如果存在，这个词已经存在，且文档号也存在，这个词不用处理了
					isExist := false
					for _, v := range docs {
						if v == xkcdResult.Num {
							isExist = true
							break
						}
					}
					if !isExist {
						//加入int[]
						docs = append(docs, xkcdResult.Num)
						indexmap[sp] = docs
					}
				}
			}
		}
		for k, v := range indexmap {
			log.WithFields(log.Fields{
				"key":   k,
				"value": v,
			}).Trace("finished who comic process")
		}
		//取下一篇
		i++
	}

	afterUpdateContext := log.WithFields(log.Fields{
		"indexNum": i,
		"indexMap": indexmap,
	})
	//持久化到磁盘
	jsonMap, err := json.Marshal(indexmap)
	if err != nil {
		afterUpdateContext.WithError(err).Fatal("json.Marshal(indexmap) err")
	}
	if err = writeIndexIntoFiles(i, string(jsonMap)); err != nil {
		afterUpdateContext.WithError(err).Fatal("indexnum and indexmap write into files err")
	}
	afterUpdateContext.Info("indexnum and indexmap write into files success")
}

func writeIndexIntoFiles(indexNum int, jsonMap string) error {
	err := os.WriteFile(INDEXNUM_FNAME, []byte(strconv.Itoa(indexNum)), 0666)
	if err != nil {
		return err
	}

	err = os.WriteFile(REALINDEX_FNAME, []byte(jsonMap), 0666)
	if err != nil {
		return err
	}
	return nil
}

// 每次分词只取[[]]内部括起来的所有word，word之间使用空格区分
func scanScriptLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	//找到第一个字母
	start := -1
	for i := 0; i < len(data); i++ {
		if data[i] == '[' {
			start = i
			break
		}
	}

	if start == -1 {
		return len(data), nil, nil
	}

	//找到第一个]。返回区间内的部分
	for i := start + 1; i < len(data); i++ {
		if data[i] == ']' {
			return i + 2, data[(start + 2):i], nil
		}
	}

	//如果没有找到，则继续申请
	return start, nil, nil
}

//搜索算法
//比如 1 2 3 4 4个(k个)[]int取交集，应该进行 1 2 --》结果1 结果1 3 --》结果2  结果2 4--》最终结果
// 所以外层循环为k-1次，内层比较算法，使用额外的map以减少时间复杂度，核心算法如下：
// 1. 两者较长的int[]，遍历存入map
// 2. 较短的遍历，在map中取key判断，如果存在则记录结果到较短[]int（原地算法，不额外申请内存）
//时间复杂度为 O(k*N) = O(N) k为用户传入的关键词个数，N为最长的int[]长度，空间复杂度为O(N)
func doSearch() []int {
	kwArr := strings.Split(keywords, " ")

	//检查是否均不为空，有一个为空，则无结果
	for _, v := range kwArr {
		if _, ok := indexmap[v]; !ok {
			return []int{}
		}
	}

	lastResultArr := indexmap[kwArr[0]]
	for i := 1; i < len(kwArr); i++ {
		//两个int[]都不是nil才进行比较判断，否则选择不是nil的那个
		lastResultArr = doRealSearch(lastResultArr, indexmap[kwArr[i]])
		if len(lastResultArr) == 0 {
			return lastResultArr
		}
	}
	return lastResultArr
}

func doRealSearch(arr1, arr2 []int) []int {
	longerArr := arr1
	shorterArr := arr2
	if len(arr2) < len(arr1) {
		longerArr = arr2
		shorterArr = arr1
	}
	r := make([]int, 0, len(shorterArr))
	//遍历存入临时map
	tm := map[int]bool{}
	for _, v := range longerArr {
		tm[v] = true
	}

	for _, v := range shorterArr {
		if tm[v] {
			r = append(r, v)
		}
	}

	return r
}

//展示给user
func showToUser(r []int) {
	fmt.Println("Search Result:")
	start := 0
	for {
		//展示五条
		s, n := doPrint(start, r)

		for _, v := range s {
			fmt.Print(v)
		}

		if n < 5 || (start+n) == len(r) {
			fmt.Println("print finished bye~")
			return
		}

		fmt.Println("see next page results?(y n) ")
		var sm string
		fmt.Scan(&sm)
		switch sm {
		case "y":
			start += 5
		case "n":
			fmt.Println("bye~")
			return
		default:
			fmt.Println("wrong param,bye~")
			return
		}
	}
}

//传入开始索引和全部的结果切片
//返回 应打印的字符串数组 和 实际打印条数
func doPrint(start int, r []int) (s []string, printSize int) {
	//展示五条
	s = []string{}
	i := start
	for ; i < start+5; i++ {
		if len(r)-1 < i {
			//不够了，无法打印5条
			return s, i - start + 1
		}
		s = append(s, fmt.Sprintf("%d: %s\n", (i+1), URL_PREFIX+strconv.Itoa(r[i])+URL_SUFFIX))
	}
	return s, 5
}
