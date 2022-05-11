package service

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"

	"github.com/akhilrex/podgrab/db"
	"github.com/akhilrex/podgrab/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func CreatePodcast(title string) (db.Podcast, error){
	var podcast db.Podcast
	err := db.GetPodcastByTitleAndAuthor(title, "", &podcast)
	if errors.Is(err, gorm.ErrRecordNotFound){
		podcast = db.Podcast{
			Title: title,
			Summary: title,
			Author: "",
			Image: "",
			URL: "",
		}

		err = db.CreatePodcast(&podcast)
	}
	// Already add; do nothing
	return podcast, &model.PodcastAlreadyExistsError{Url: ""}
}

func fillInPodcastItem(folder string, filename string, filesize int64, podcast *db.Podcast) db.PodcastItem{
	itemTitle := strings.Split(filename,".")[0]
	downloadPath := path.Join(folder, filename)
	imagePath := path.Join(folder, "folder.jpg")
	podcastItem := db.PodcastItem{
		PodcastID:	podcast.ID,
		Title:     	itemTitle,
		Summary:   	itemTitle,
		EpisodeType: 	"full",
		Duration:	0,
		PubDate:	time.Now(),
		FileURL:	"",
		GUID:		"",
		Image:		"",
		DownloadDate:	time.Now(),
		DownloadPath:	downloadPath,
		DownloadStatus:	db.Downloaded,
		IsPlayed:	false,
		BookmarkDate:	time.Time{},
		LocalImage:	imagePath,
		FileSize:	filesize,
	}
	return podcastItem
}

func getPodcastItemByTitle(title string, podcastItem *db.PodcastItem) error {
	result := db.DB.Preload(clause.Associations).First(&podcastItem, "title=?", title)
	return result.Error
}

func CreatePodcastItems(podcast *db.Podcast) {

	dataPath := os.Getenv("DATA")
	title := podcast.Title

	folder := path.Join(dataPath, title)

	files, err := ioutil.ReadDir(folder)
	if err != nil {
		fmt.Println("ReadDir: ", err)
	}

	for _, fInfo := range files {
		isdir:= fInfo.IsDir()
		filename := fInfo.Name()
		filesize := fInfo.Size()
		if !isdir &&
		strings.HasSuffix(filename, ".m4a") ||
		strings.HasSuffix(filename, ".mp3") ||
		strings.HasSuffix(filename, ".aac") ||
		strings.HasSuffix(filename, ".ogg") ||
		strings.HasSuffix(filename, ".opus") ||
		strings.HasSuffix(filename, ".wav") ||
		strings.HasSuffix(filename, ".flac") {
			var podcastItem db.PodcastItem
			err := getPodcastItemByTitle(strings.Split(filename,".")[0], &podcastItem)
			if errors.Is(err, gorm.ErrRecordNotFound){
				podcastItem = fillInPodcastItem(folder,filename, filesize, podcast)
				db.CreatePodcastItem(&podcastItem)
			}
			// Already add; do nothing
		}
	}
	return
}

func updatePodcastImage(podcast *db.Podcast) {
	setting := db.GetOrCreateSetting()
	baseURL := setting.BaseUrl
	imageURL := fmt.Sprintf("%s/podcasts/%s/image",baseURL,podcast.ID)
	db.DB.Model(db.Podcast{}).Where("id=?", podcast.ID).Update("image", imageURL)
//	db.DB.Model(db.PodcastItem{}).Where("podcast_id=?", podcast.ID).Update("image",imageURL)
	return
}

func intersect(slice1, slice2 [] string) []string{
	m := make(map[string]int)
	nn:= make([]string,0)
	for _,v := range slice1{
		m[v]++
	}

	for _,v := range slice2{
		times,_:=m[v]
		if times == 1{
			nn=append(nn,v)
		}

	}
	return nn
}

//slice1\slice2
func difference(slice1, slice2 [] string) []string{
	m := make(map[string]int)
	nn := make([]string,0)
	inter := intersect(slice1,slice2)
	for _,v :=range inter{
		m[v]++
	}

	for _,value := range slice1 {
		times,_:=m[value]
		if times == 0{
			nn=append(nn,value)
		}
	}
	return nn
}

func CheckNewFolders() {
	var podcasts []db.Podcast
	var indatabase []string
	var infilesystem []string

	db.DB.Preload(clause.Associations).Find(&podcasts)
	for _, podcast := range podcasts{
		indatabase = append(indatabase,strings.Trim(podcast.Title," "))
	}

	dataPath := os.Getenv("DATA")
	files, err := ioutil.ReadDir(dataPath)
	if err != nil {
		fmt.Println("ReadDir: ", err)
	}

	for _, fInfo := range files {
		isdir:= fInfo.IsDir()
		filename := fInfo.Name()
		if isdir {
			infilesystem = append(infilesystem, filename)
		}
	}

	diff := difference(infilesystem,indatabase)
	for _,name := range diff{
		fmt.Println(name)
		podcast, _ := CreatePodcast(name)
		CreatePodcastItems(&podcast)
		updatePodcastImage(&podcast)
	}
}
