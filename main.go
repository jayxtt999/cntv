package main

import (
	"flag"
	"fmt"
	"github.com/jayxtt999/go-ncpa-classic/service"
	"os"
	"strconv"
)

var (
	ncpaUrl      string
	list      bool
	down      bool
	url      string
	page      int
	num      int
)
func init() {
	flag.StringVar(&ncpaUrl, "u", "", "单个视频地址，url中带有VID")
	flag.BoolVar(&list,"list", true, "获取视频列表")
	flag.IntVar(&page,"page", 0, "当前页")
	flag.IntVar(&num,"num", 20, "每页大小")
	flag.BoolVar(&down,"d", false, "是否下载")

}

func main() {
	flag.Parse()
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("[error]", r)
			os.Exit(-1)
		}
	}()

	if ncpaUrl!="" && down {
		err:= service.DownVideo(ncpaUrl)
		if err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	if list {
		p := strconv.Itoa(page)
		n := strconv.Itoa(num)
		playListData,err := service.GetVideoList(p,n)
		if err != nil {
			panic(err)
		}
		for _,play := range playListData{
			fmt.Println(play.Title,"-",play.Time)
			if down{
				fmt.Println("download start：",play.Title)
				vidUrl,err:= service.GetVidUrl(play.Id)
				if err != nil {
					panic(err)
				}
				err = service.DownVideo(vidUrl)
				if err != nil {
					panic(err)
				}
			}
		}
		os.Exit(0)
	}


}

func panicParameter(name string) {
	panic("parameter '" + name + "' is required")
}