package rbdbgen

import (
	"math"
	"os"
	"path"
	"path/filepath"
	"rbdbtools/pkg/cache"
	"rbdbtools/pkg/database"
	"rbdbtools/pkg/logger"
	"rbdbtools/tools"
	"regexp"
	"time"
)

const (
	internalCacheName = "internalTags.json"
	externalCacheName = "externalTags.json"
	getTrackIncrement = 10000
)

var (
	matchSong = regexp.MustCompile(`.*\.(mp3|m4a|flac)`)
	log       = logger.New()
)

func timer(t time.Time) {
	log.Infof("Created database in %s", time.Since(t))
}

func Rbdbgen(bigEndian bool, targetDir string, internalTrackDir string, externalTrackDir string) {
	err := os.MkdirAll(targetDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Starting...")
	defer timer(time.Now())

	db := database.New(bigEndian)

	oldCacheSize, newCacheSize := 0, 0
	if internalTrackDir != "" {
		o, n := loadTracksIntoDB(loadTracksIntoDBParams{
			tracksPath:    internalTrackDir,
			targetDir:     targetDir,
			cacheLocation: internalCacheName,
			external:      false,
			database:      &db,
		})
		oldCacheSize += o
		newCacheSize += n
	}

	if externalTrackDir != "" {
		o, n := loadTracksIntoDB(loadTracksIntoDBParams{
			tracksPath:    externalTrackDir,
			targetDir:     targetDir,
			cacheLocation: externalCacheName,
			external:      true,
			database:      &db,
		})
		oldCacheSize += o
		newCacheSize += n
	}

	endian := "little"
	if bigEndian {
		endian = "big"
	}
	log.Infof("Saving database in %s endian format...", endian)

	err = db.Save(targetDir)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Cache increased by %d, cache size is now %d", newCacheSize-oldCacheSize, newCacheSize)

	for k, v := range db.Entries() {
		log.Infof("Database has %d %s entries, a recorded size of %d bytes, and an actual size of %d bytes", v.Header.Entries, k, v.Header.Size, v.Size)
	}

	log.Infof("DB Size: %s", tools.BytesToFormalSize(db.Size()))
}

type loadTracksIntoDBParams struct {
	tracksPath    string
	targetDir     string
	cacheLocation string
	external      bool
	database      *database.Database
}

func loadTracksIntoDB(params loadTracksIntoDBParams) (int, int) {
	location := ""
	locationName := "internal"
	if params.external {
		location = "<microSD1>"
		locationName = "external"
	}

	log.Infof("Loading %s cache...", locationName)
	c, err := cache.New(path.Join(params.targetDir, params.cacheLocation), params.tracksPath, location)
	if err != nil {
		log.Error(err)
	}
	oldSize := c.Size()

	log.Infof("Adding %s storage tracks from '%s'", locationName, params.tracksPath)
	fileList, err := getTracks(params.tracksPath)
	if err != nil {
		log.Error(err)
	}

	log.Info("Getting tags for tracks...")
	tagList, err := getTags(fileList, c)

	log.Info("Adding tracks to database...")
	params.database.Add(tagList...)

	log.Info("Saving cache...")
	err = c.Save()
	if err != nil {
		log.Error(err)
	}
	return oldSize, c.Size()
}

func getTracks(root string) ([]string, error) {
	i := 0
	tracks := make([]string, 0)
	err := filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error(err)
			return nil
		} else if info.IsDir() || !matchSong.MatchString(p) {
			return nil
		}

		i++
		if i%100 == 0 {
			log.Infof("Found %d tracks...", i)
		}

		tracks = append(tracks, p)

		return nil
	})
	if i%100 != 0 {
		log.Infof("Found %d tracks...", i)
	}
	return tracks, err
}

func getTags(tracks []string, cache cache.Cache) ([]database.Track, error) {
	t := make([]database.Track, 0)
	for i := 0; i < len(tracks); {
		end := int(math.Min(float64(len(tracks)), float64(i+getTrackIncrement)))

		log.Infof("Getting tags for %d tracks...", end-i)
		tags, err, cached := cache.Add(tracks[i:end]...)
		t = append(t, tags...)

		log.Infof("%d tags were read from disk", (end-i)-cached)
		if err != nil {
			log.Warningf("%d tags may not have been read", cached)
			return t, err
		} else {
			log.Infof("%d tags were already cached", cached)
		}

		i = end
	}
	return t, nil
}
