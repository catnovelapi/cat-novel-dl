package main

import (
	"flag"
	"fmt"
	"github.com/catnovelapi/cat"
	"github.com/tidwall/gjson"
	"log"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

type Person struct {
	bookId      string
	bookName    string
	saveDir     string
	loginToken  string
	account     string
	chaptersDir string
	client      *cat.Ciweimao
}

func init() {
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Println("打开日志文件失败：", err.Error())
		return
	}
	log.SetOutput(file)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func (person *Person) download(bookInfo gjson.Result) {
	person.bookId = bookInfo.Get("book_id").String()
	person.bookName = bookInfo.Get("book_name").String()
	if person.bookName == "" {
		log.Fatalf("bookId:%s,获取书籍信息失败", person.bookId)
	} else {
		fmt.Println("开始下载:", person.bookName)
	}
	person.chaptersDir = person.newFile(path.Join(person.saveDir, person.bookName, "chapters"))
	wg := sync.WaitGroup{}
	for _, chapter := range person.client.NewCatalogByBookIDApi(person.bookId) {
		wg.Add(1)
		go func(chapter gjson.Result, wg *sync.WaitGroup) {
			defer wg.Done()
			chapterId := chapter.Get("chapter_id").String()
			if chapter.Get("auth_access").String() == "1" &&
				!person.exists(path.Join(person.chaptersDir, chapterId+".txt")) {

				command := person.client.ChapterCommandApi(chapterId).Get("data.command").String()
				if content := person.client.ChapterInfoApi(chapterId, command); content != "" {
					fmt.Println(chapter.Get("chapter_title").String(), " 下载完成!\r")
					person.writeFile(path.Join(person.chaptersDir, chapterId+".txt"), content)
				}
			} else {
				log.Println("chapterId:", chapterId, "warning:该章节需要付费或者已经下载过了")
			}
		}(chapter, &wg)
	}
	wg.Wait()
	fmt.Println(person.bookName, "下载完毕!")
}
func (person *Person) outFile() {
	p := path.Join(path.Dir(person.chaptersDir), person.bookName+".txt")
	if err := os.Truncate(p, 0); err != nil {
		fmt.Println("清空文件失败:", err)
	}
	file, err := os.OpenFile(p, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		fmt.Println("无法打开文件:", err)
		return
	}
	defer file.Close()
	for _, chapter := range person.client.NewCatalogByBookIDApi(person.bookId) {
		chapterId := chapter.Get("chapter_id").String()
		if person.exists(path.Join(person.chaptersDir, chapterId+".txt")) {
			content, err := os.ReadFile(path.Join(person.chaptersDir, chapterId+".txt"))
			if err != nil {
				log.Println("读取文件失败:", err)
			} else {
				_, _ = file.Write([]byte(chapter.Get("chapter_title").String() + "\n" + string(content) + "\n\n"))
			}
		}
	}
	fmt.Printf("\n<%v>合并完毕,三秒后自动关闭终端", person.bookName)
	time.Sleep(3000 * time.Second)
}

func (person *Person) newFile(name string) string {
	_, err := os.Stat(name)
	if err != nil {
		err = os.MkdirAll(name, os.ModePerm)
		if err != nil {
			fmt.Println("创建文件夹失败:", err)
		}
	}
	return name
}

func (person *Person) exists(name string) bool {
	_, err := os.Stat(name)
	return err == nil || os.IsExist(err)
}

func (person *Person) writeFile(name string, content string) {
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		fmt.Println("写入文件失败:", err)
	}
}

func main() {
	var person Person
	flag.StringVar(&person.bookId, "d", "", "add bookid to download book")
	flag.StringVar(&person.saveDir, "o", "books", "save file name")
	flag.StringVar(&person.bookName, "s", "", "Search book name to download book")
	flag.StringVar(&person.account, "a", "", "set account")
	flag.StringVar(&person.loginToken, "t", "", "set login token")
	flag.Parse()
	if person.loginToken != "" && person.account != "" {
		if len(person.loginToken) != 32 {
			log.Fatalf("token长度不正确,必须为32位")
		} else if !strings.Contains(person.account, "书客") {
			log.Fatalf("账号格式不正确,必须以`书客`开头")
		}
		person.client = cat.NewCiweimaoClient(cat.Account(person.account), cat.LoginToken(person.loginToken))
	} else {
		log.Fatalf("账号或者token不能为空,请使用 -a 书客12345 -t 32位token")
	}

	if person.bookId != "" {
		person.download(person.client.BookInfoApi(person.bookId).Get("data.book_info"))
		person.outFile()
	} else if person.bookName != "" {
		resultArray := person.client.SearchByKeywordApi(person.bookName, 0).Get("data.book_list").Array()
		for i, result := range resultArray {
			fmt.Println("INDEX", i, "\t\tBOOK_NAME", result.Get("book_name").String())
		}
		var bookIndex int
		for {
			fmt.Printf("请输入要下载的书籍序号:")
			fmt.Scanln(&bookIndex)
			if bookIndex < len(resultArray) {
				person.download(resultArray[bookIndex])
				person.outFile()
			}
		}
	} else {
		fmt.Println("请输入参数,使用 -h 查看帮助")
	}
}
