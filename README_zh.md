goasync
=========================

[English](https://github.com/qjpcpu/goasync/blob/master/README.md)

`goasync`受node.js的一个著名库[async](https://github.com/caolan/async)启发而来，旨在解决golang并发编程的共性问题:goroutine的使用。

### Install

```
# 安装依赖
go get github.com/qjpcpu/goasync
# 引用库
import "github.com/qjpcpu/goasync"
```

### 使用举例

比如，有时我们需要并行执行多个任务，那么我们可能就需要开启多个go routine来处理，当所有任务完成时，检查并处理执行结果，这种场景很常见，实现也很简单。既然这样，其实就说明这种模式可以归纳提炼，这就是`Parallel`接口解决的问题。

#### `func Parallel(functions ...TaskHandler) (async *Async, err error)`

该函数接受不定参数`TaskHandler`，该参数其实是一个函数，其函数签名为`func(Cb, ResultSet)`。

第一个参数`Cb`为回调函数，必须在传入的用户函数中调用; 第二个参数`ResultSet`是上游任务执行结果,这`Parallel`场景下一般不需要用到。

```go
package main

import (
	"github.com/qjpcpu/goasync"
	"log"
	"time"
)

func main() {
	async, err := goasync.Parallel(
		func(cb goasync.Cb, ar goasync.ResultSet) {
			log.Println("task 0 started, sleep 5 seconds...")
			time.Sleep(time.Second * 5)
			log.Println("task 0 done")
			cb(nil, nil)
		},
		func(cb goasync.Cb, ar goasync.ResultSet) {
			log.Println("task 1 started, calculate 2 * 3 = ?")
			time.Sleep(time.Second * 2)
			res := 2 * 3
			log.Println("task 1 done,return result.")
			cb(res, nil)
		},
	)
	if err != nil {
		log.Fatalln("Failed to create parallel task")
	}
	err = async.Run()
	if err != nil {
		log.Fatalln("An error occur in certain goasync task")
	}
	var data int
    // fetch results by task index, note: the index type is string not integer
	async.GetResult("1").Data(&data)
	log.Printf("get task 1 result: 2 * 3 = %v\n", data)
}
```

输出结果为:

```
2016/04/09 22:04:02 task 0 started, sleep 5 seconds...
2016/04/09 22:04:02 task 1 started, calculate 2 * 3 = ?
2016/04/09 22:04:04 task 1 done,return result.
2016/04/09 22:04:07 task 0 done
2016/04/09 22:04:07 get task 1 result: 2 * 3 = 6
```


#### `func Auto(flows map[string]*Task) (async *Async, err error)`

而`Auto`接口对应的场景就更多了:

比如，你希望从网上搜索下载一张图片，然后交给你的机器人去修剪图片再保存到本地磁盘的某个文件夹，而当你下载完图片后，不需要等待机器人做图片处理就可以打开你的电话薄到肯德基给自己订份外卖吃;当所有的事情完成后，就可以下班了。

这其实就是解决拓扑排序的相关问题，然而每次去为这样的事情写代码也挺重复的，不外乎就是解决多个go routine的执行顺序及相互通信。 使用`Auto`接口，可以让你仅定义任务的依赖关系及实现逻辑，协同的事情就交给`goasync`吧。

```go
package main

import (
	"github.com/qjpcpu/goasync"
	"log"
	"time"
)

func main() {
	flows := map[string]*goasync.Task{
		"download-image": &goasync.Task{
			Handler: func(cb goasync.Cb, ar goasync.ResultSet) {
				url := "http://somewhere.com/flower.jpeg"
				log.Printf("[download-image]\tDownloading %s ...\n", url)
				time.Sleep(time.Second * 2)
				cb("flower.jpeg", nil)
			},
		},
		"resize-image": &goasync.Task{
			Dep: []string{"download-image"},
			Handler: func(cb goasync.Cb, ar goasync.ResultSet) {
				var filename string
				ar.Get("download-image").Data(&filename)
				log.Printf("[resize-image]\tThe robot now can load %s & resize it...\n", filename)
				time.Sleep(time.Second * 3)
				fullpath := "/my-folder/" + filename
				cb(fullpath, nil)
			},
		},
		"save-image": &goasync.Task{
			Dep: []string{"resize-image"},
			Handler: func(cb goasync.Cb, ar goasync.ResultSet) {
				var fullpath string
				ar.Get("resize-image").Data(&fullpath)
				time.Sleep(time.Second * 2)
				log.Printf("[save-image]\tSave image to %s...\n", fullpath)
				time.Sleep(time.Second * 1)
				cb(nil, nil)
			},
		},
		"search-phonebook": &goasync.Task{
			Dep: []string{"download-image"},
			Handler: func(cb goasync.Cb, ar goasync.ResultSet) {
				log.Println("[search-phonebook]\tFind phonebook can look for the phone number of KFC...")
				time.Sleep(time.Second * 3)
				number := "4008-517-517"
				log.Printf("[search-phonebook]\tGot KFC number:%s\n", number)
				cb(number, nil)
			},
		},
		"make-order": &goasync.Task{
			Dep: []string{"search-phonebook"},
			Handler: func(cb goasync.Cb, ar goasync.ResultSet) {
				var number string
				ar.Get("search-phonebook").Data(&number)
				log.Printf("[make-order]\tCall %s and order my launch...\n", number)
				time.Sleep(time.Second * 1)
				log.Println("[make-order]\tOrder OK, enjoy launch...")
				cb(nil, nil)
			},
		},
		"off-work": &goasync.Task{
			Dep: []string{"make-order", "save-image"},
			Handler: func(cb goasync.Cb, ar goasync.ResultSet) {
				time.Sleep(time.Second * 1)
				log.Println("[off-work]\tSave image done & finish my launch, off work ^_^")
				cb(nil, nil)
			},
		},
	}
	async, _ := goasync.Auto(flows)
	async.Run()
	log.Println("All tasks done.")
}
```

The output should be:

```
2016/04/09 22:35:00 [download-image]    Downloading http://somewhere.com/flower.jpeg ...
2016/04/09 22:35:02 [search-phonebook]  Find phonebook can look for the phone number of KFC...
2016/04/09 22:35:02 [resize-image]      The robot now can load flower.jpeg & resize it...
2016/04/09 22:35:05 [search-phonebook]  Got KFC number:4008-517-517
2016/04/09 22:35:05 [make-order]        Call 4008-517-517 and order my launch...
2016/04/09 22:35:06 [make-order]        Order OK, enjoy launch...
2016/04/09 22:35:07 [save-image]        Save image to /my-folder/flower.jpeg...
2016/04/09 22:35:09 [off-work]          Save image done & finish my launch, off work ^_^
2016/04/09 22:35:09 All tasks done.
```

### 其他

欢迎`pull request`。
