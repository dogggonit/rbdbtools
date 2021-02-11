package taglib

// #cgo pkg-config: taglib
// #cgo LDFLAGS: -ltag
// #include <stdlib.h>
// #include "tags.h"
import "C"

import (
	"errors"
	"os"
	"rbdbtools/pkg/database"
	"unsafe"
)

type TagLib struct {
	artist      string
	album       string
	genre       string
	title       string
	filename    string
	composer    string
	comment     string
	albumArtist string
	grouping    string
	year        int32
	disc        int32
	track       int32
	bitrate     int32
	length      int32
	mtime       int32
}

var GetTrackError = errors.New("failed to get tracks")

func New(filename string) (database.MetaData, error) {
	cs := C.CString(filename)
	defer C.free(unsafe.Pointer(cs))

	t := C.getTrack(cs)
	if t == nil {
		return &TagLib{}, GetTrackError
	}
	defer C.freeTrack(t)

	return cTrackToTrack(t)
}

func NewTracks(filenames ...string) ([]database.MetaData, error) {
	cFilenames, cleanup := toCharStarStar(filenames)
	defer cleanup()

	cTracks := C.getTracks(cFilenames, C.size_t(len(filenames)))
	if cTracks == nil {
		var t [0]database.MetaData
		return t[:], GetTrackError
	}
	defer C.freeTracks(cTracks)

	numTracks := int(cTracks.size)
	tracks := make([]database.MetaData, numTracks)
	for i, e := range ((*[1 << 30]*C.struct_track)(unsafe.Pointer(cTracks.tracks)))[:numTracks:numTracks] {
		var err error
		tracks[i], err = cTrackToTrack(e)
		if err != nil {
			return tracks[:i], err
		}
	}

	return tracks, nil
}

// toCharStarStar converts a go slice of strings to a c array of c strings
func toCharStarStar(a []string) (**C.char, func()) {
	cA := C.malloc(C.size_t(len(a)) * C.size_t(unsafe.Sizeof(uintptr(0))))

	cAS := ((*[1 << 30]*C.char)(cA))[:len(a):len(a)]
	for i, s := range a {
		cAS[i] = C.CString(s)
	}

	return (**C.char)(cA), func() {
		for i := range cAS {
			C.free(unsafe.Pointer(cAS[i]))
		}
		C.free(cA)
	}
}

// cTrackToTrack converts a struct track to a database.MetaData
func cTrackToTrack(track *C.struct_track) (database.MetaData, error) {
	if track == nil {
		return &TagLib{}, errors.New("cannot get info from null track")
	}

	filename := C.GoString(track.filename)

	mtime, err := getMTime(filename)
	if err != nil {
		return &TagLib{}, err
	}

	t := &TagLib{
		artist:      C.GoString(track.artist),
		album:       C.GoString(track.album),
		genre:       C.GoString(track.genre),
		title:       C.GoString(track.title),
		filename:    filename,
		composer:    C.GoString(track.composer),
		comment:     C.GoString(track.comment),
		albumArtist: C.GoString(track.albumArtist),
		grouping:    C.GoString(track.grouping),
		year:        int32(track.year),
		disc:        int32(track.disc),
		track:       int32(track.track),
		bitrate:     int32(track.bitrate),
		length:      int32(track.length),
		mtime:       mtime,
	}

	if t.AlbumArtist() == "" || t.AlbumArtist() == "<Untagged>" {
		t.albumArtist = t.Artist()
	}

	if t.Grouping() == "" || t.Grouping() == "<Untagged>" {
		t.grouping = t.Title()
	}

	return t, nil
}

// getMTime gets the mtime of a file, in unix milliseconds, as an int32
func getMTime(filename string) (int32, error) {
	file, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	mtime := file.ModTime()

	return int32(mtime.Unix()), nil
}
