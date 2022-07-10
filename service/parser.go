package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type VideoListByAlbumIdNewDataRes struct{
	Data VideoListByAlbumIdNewData `json:"data"`
}
type VideoListByAlbumIdNewData struct{
	Total int `json:"total"`
	List []VideoListByAlbumIdNewDataList `json:"list"`

}
type VideoListByAlbumIdNewDataList struct {
	Url string `json:"url"`
	Id string `json:"id"`
}

type playList struct{
	Error bool `json:"error"`
	TotalNums int `json:"totalnums"`
	Data []playListData `json:"data"`

}

type playListData struct {
	Title string `json:"DRETITLE"`
	Link string `json:"PAGELINK"`
	Time string `json:"PLAYTIME"`
	Id string `json:"SOURCEDB_ID"`
}


type videoListList []videoListListData

type videoListByAlbumIdNewData struct {
	Total int `json:"total"`
	List  videoListList `json:"list"`
}

type videoListListData struct{
	Id string `json:"id"`
	Title string `json:"title"`
	Guid string `json:"guid"`
	Length string `json:"length"`
	DownUrl string
}

type videoListByAlbumIdNew struct {
	Data videoListByAlbumIdNewData `json:"data"`
}
type httpVideoInfo struct {
	HlsUrl string `json:"hls_url"`
}
func  getVid(url string) (string, error) {

	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	html := string(body)
	//fmt.Println(html)
	r,_:=regexp.Compile(`videotvCodes\=\"(VID.*)\"\;`)
	matchRes := r.MatchString(html)
	if !matchRes{
		return "", fmt.Errorf("get vid err")
	}
	matched := r.FindAllStringSubmatch(html, 1)
	vid := matched[0][1]
	if vid==""{
		return "", fmt.Errorf("get vid error")
	}

	//r,_:=regexp.Compile(`^\/\d+\/\d+\/\d+\/(.*)\.shtml$`)
	//matchRes := r.MatchString(path)
	//if !matchRes{
	//	return "", fmt.Errorf("url path non compliance with rules")
	//}
	//matched := r.FindAllStringSubmatch(path, 1)
	//vid := matched[0][1]
	//if vid==""{
	//	return "", fmt.Errorf("get vid error")
	//}
	return vid,nil
}

func getRes(url string) ([]byte, error) {
	//fmt.Println(url)
	response, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	return  body,err
}

func getVideoListList(vid string) (videoListList, error) {
	//https://api.cntv.cn/NewVideo/getVideoListByAlbumIdNew?id=VIDADBHD1F0vSRUn1GfKW8O4211102&serviceId=chncpa&p=1&n=100&mode=2&pub=2&sort=asc&callback=jQuery17205304378474440867_1638798335030&_=1638798336005
	//https://vdn.apps.cntv.cn/api/getHttpVideoInfo.do?pid=9a8f6ac48cb74e7fabd4c0c82590e41c
	timeUnix:=time.Now().UnixNano() / 1e6
	url := "https://api.cntv.cn/NewVideo/getVideoListByAlbumIdNew?id="+vid+"&serviceId=chncpa&p=1&n=100&mode=2&pub=2&sort=asc&_="+strconv.FormatInt(timeUnix,10)
	body, err := getRes(url)
	if err != nil {
		panic(err)
	}
	videoList := videoListByAlbumIdNew{}
	jsonErr := json.Unmarshal(body, &videoList)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	if videoList.Data.Total<=0{
		return nil, nil
	}
	return  videoList.Data.List,nil
}

func getHost(link string)(string,error)  {
	u, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	return u.Host, err
}

func GetDownList(ncpaUrl string) (videoListList, error) {


	host, err := getHost(ncpaUrl)
	if err != nil {
		return nil, err
	}
	if host != "www.ncpa-classic.com"{
		return nil, fmt.Errorf("url host  must be www.ncpa-classic.com")
	}
	vid,err := getVid(ncpaUrl)
	if err != nil {
		return nil, err
	}
	//fmt.Println("VID:",vid)
	videoListList,err:=getVideoListList(vid)
	//s := *videoListList
	if err != nil {
		return nil, err
	}
	if videoListList == nil {
		return nil, err
	}
	videoInfoUrl := "https://vdn.apps.cntv.cn/api/getHttpVideoInfo.do?pid=guid"
	//urlListMaps = make(map[int]string)
	for i,item := range videoListList{
		_url  := strings.Replace(videoInfoUrl, "guid", item.Guid, -1)
		body, err := getRes(_url)
		if err != nil {
			panic(err)
		}
		httpVideo := httpVideoInfo{}
		jsonErr := json.Unmarshal(body, &httpVideo)
		if jsonErr != nil {
			log.Fatal(jsonErr)
		}
		if httpVideo.HlsUrl != ""{
			mainUrl := httpVideo.HlsUrl
			mainFile , err := getRes(mainUrl)
			if err != nil {
				panic(err)
			}
			countSplit := strings.Split(string(mainFile), "\n")
			lastRow := countSplit[len(countSplit)-2]
			lastRow = strings.Replace(lastRow, "\n", "", -1)
			if lastRow != "#EXT-X-ENDLIST"{
				domain, err := getHost(mainUrl)
				if err != nil {
					return nil, err
				}
				mainUrl = "https://"+domain+"/"+lastRow
			}
			videoListList[i].DownUrl = mainUrl
		}else{
			videoListList[i].DownUrl = ""
		}

	}
	return  videoListList,nil
}

func  GetVideoList(page string,num string) ([]playListData, error) {
	url := "https://so.cntv.cn/sapi/clmusic/playlist_search.php?page="+page+"&num="+num+"&theme=&format=json&year="
	body, err := getRes(url)
	if err != nil {
		panic(err)
	}
	playList := playList{}
	jsonErr := json.Unmarshal(body, &playList)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	if playList.Error{
		return nil, nil
	}
	playListData := playList.Data
	return  playListData,nil
}

func DownVideo(ncpaUrl string) (err error) {

	list,err := GetDownList(ncpaUrl)
	if err != nil {
		return  err
	}
	if list == nil {
		return  fmt.Errorf("list is nil")
	}
	mainName := ""
	for k,item := range list{
		//go func(title string,url string) {
		fmt.Println("start:"+item.Title)
		fmt.Println("DownUrl:"+item.DownUrl)
		if k == 0 {
			mainName = item.Title
		}
		output:="./data/"+mainName
		downloader, err := NewTask(output, item.DownUrl)
		if err != nil {
			return  err
		}
		if err := downloader.Start(item.Title,100); err != nil {
			return  err
		}
		//}(item.Title,item.DownUrl)

	}
	return nil
}

func GetVidUrl(Id string)(url string,err error)  {

	url = "https://api.cntv.cn/NewVideo/getVideoListByAlbumIdNew?id="+Id+"&serviceId=chncpa&p=1&n=100&mode=2&pub=2&sort=asc"
	body, err := getRes(url)
	if err != nil {
		panic(err)
	}
	newDataRes := VideoListByAlbumIdNewDataRes{}
	jsonErr := json.Unmarshal(body, &newDataRes)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	if newDataRes.Data.Total<=0{
		return "", nil
	}
	list := newDataRes.Data.List
	return  list[0].Url,nil

}