goasync
=========================

[中文](https://github.com/qjpcpu/goasync/blob/master/README_zh.md)

`goasync` is inspired by the cool Node.js lib [async](https://github.com/caolan/async), I hope it makes multiple go routines programming easier.

### Install

```
# add goasync as your dependency
go get github.com/qjpcpu/goasync
# add goasync to your go source code
import "github.com/qjpcpu/goasync"
```

### Usage

#### `func Parallel(functions ...TaskHandler) (async *Async, err error)`

Run multiple tasks(`go routines`) parallel and wait them all.

And the `TaskHandler` must match the signature `func(Cb, ResultSet)`.

The parameter `Cb` is a callback function, which must be get called. If no error happens, the `Cb` should be called like `cb(data,nil)`. the `data` can be passed out.

The others parameters are upstream tasks results.

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

The output would be

```
2016/04/09 22:04:02 task 0 started, sleep 5 seconds...
2016/04/09 22:04:02 task 1 started, calculate 2 * 3 = ?
2016/04/09 22:04:04 task 1 done,return result.
2016/04/09 22:04:07 task 0 done
2016/04/09 22:04:07 get task 1 result: 2 * 3 = 6
```


#### `func Auto(flows map[string]*Task) (async *Async, err error)`

Consider this scenario: you want download an image first, then your robot can auto resize the image and store it to certain folder, and after the image downloaded, you can open your phone book and call for launch. After all, you can go off work.

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
