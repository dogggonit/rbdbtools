package cache

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"rbdbtools/pkg/database"
	"rbdbtools/pkg/taglib"
	"strings"
)

type Cache struct {
	location  string
	rootPath  string
	newPrefix string
	cache     map[string]database.Track
}

func New(cacheLocation string, rootPath string, newPrefix string) (Cache, error) {
	c := Cache{
		location:  cacheLocation,
		rootPath:  rootPath,
		newPrefix: newPrefix,
		cache:     make(map[string]database.Track),
	}
	if _, err := os.Stat(cacheLocation); err == nil {
		data, err := ioutil.ReadFile(cacheLocation)
		if err != nil {
			return Cache{}, err
		} else {
			err := json.Unmarshal(data, &c.cache)
			if err != nil {
				return c, err
			}
			return c, nil
		}
	} else if os.IsNotExist(err) {
		return c, nil
	} else {
		return Cache{}, err
	}
}

func (c *Cache) Add(filenames ...string) ([]database.Track, error, int) {
	notRead := 0

	if c.cache == nil {
		return []database.Track{}, errors.New("cache not initialized"), notRead
	}

	tracks := make([]database.Track, 0)
	newPaths := make(map[string]string)

	filenamesToRead := make([]string, 0)
	for _, e := range filenames {
		newPath := path.Join(c.newPrefix, strings.Replace(e, c.rootPath, c.newPrefix, 1))
		if !strings.Contains(newPath, "<microSD1>") && []rune(newPath)[0] != '/' {
			newPath = "/" + newPath
		}

		if t, exists := c.cache[newPath]; exists {
			tracks = append(tracks, t)
			notRead++
		} else {
			newPaths[e] = newPath
			filenamesToRead = append(filenamesToRead, e)
		}
	}

	if len(filenamesToRead) > 0 {
		tags, err := taglib.NewTracks(filenamesToRead...)
		if err != nil {
			return []database.Track{}, errors.New("failed to get tags for cache"), notRead + (len(filenames) - len(tags))
		}

		for _, t := range tags {
			t.SetFilename(newPaths[t.Filename()])
			c.cache[t.Filename()] = t
			tracks = append(tracks, t)
		}

		return tracks, nil, notRead
	}
	return tracks, nil, notRead
}

func (c *Cache) Contains(key string) bool {
	newPath := path.Join(c.newPrefix, strings.Replace(key, c.rootPath, "", 1))
	_, e := c.cache[newPath]
	return e
}

func (c *Cache) Save() error {
	j, err := json.MarshalIndent(c.cache, "", "    ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.location, j, os.ModePerm)
}

func (c *Cache) Size() int {
	return len(c.cache)
}
