## Notes for podgrab

### main cron job

In `main.go`, the `intiCron` function calls a bunch of services.

```go
func intiCron() {
	checkFrequency, err := strconv.Atoi(os.Getenv("CHECK_FREQUENCY"))
	if err != nil {
		checkFrequency = 30
		log.Print(err)
	}
	service.UnlockMissedJobs()
	//gocron.Every(uint64(checkFrequency)).Minutes().Do(service.DownloadMissingEpisodes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.RefreshEpisodes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.CheckMissingFiles)
	gocron.Every(uint64(checkFrequency) * 2).Minutes().Do(service.UnlockMissedJobs)
	gocron.Every(uint64(checkFrequency) * 3).Minutes().Do(service.UpdateAllFileSizes)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.DownloadMissingImages)
	gocron.Every(uint64(checkFrequency)).Minutes().Do(service.CheckNewFolders)
	gocron.Every(2).Days().Do(service.CreateBackup)
	<-gocron.Start()
}
```

`service.UnlockMissedJobs`: `db.UnlockMissedJobs`
`service.RefreshEpisodes`: traverse all podcasts and call `AddPodcastItems`
`service.CheckMissingFiles`: get all podcast items already downloaded, check local storage if file exists.
`service.UpdateAllFileSizes`: Update podcast items file sizes.
`DownloadMissingImages`: only download images when enabled in settings and no image info in the database.
`CreateBackup`: backup `podgrab.db` every 2 days.
`CheckNewFolders`: Check local albums and add them to the podcast.
