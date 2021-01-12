package track

// #cgo pkg-config: taglib
// #cgo LDFLAGS: -ltag
// #include <stdlib.h>
// #include "track.h"
import "C"

import (
	"encoding/json"
	"errors"
	"os"
	"unsafe"
)

type Track struct {
	Artist      string
	Album       string
	Genre       string
	Title       string
	Filename    string
	Composer    string
	Comment     string
	AlbumArtist string
	Grouping    string
	Year        uint32
	Disc        uint32
	Track       uint32
	Bitrate     uint32
	Length      uint32
	Mtime       uint32
}

var GetTrackError = errors.New("failed to get tracks")

func New(filename string) (Track, error) {
	cs := C.CString(filename)
	defer C.free(unsafe.Pointer(cs))

	track := C.getTrack(cs)
	if track == nil {
		return Track{}, GetTrackError
	}
	defer C.freeTrack(track)

	return cTrackToTrack(track)
}

func NewTracks(filenames ...string) ([]Track, error) {
	cFilenames, cleanup := toCharStarStar(filenames)
	defer cleanup()

	cTracks := C.getTracks(cFilenames, C.size_t(len(filenames)))
	if cTracks == nil {
		return []Track{}, GetTrackError
	}
	defer C.freeTracks(cTracks)

	numTracks := int(cTracks.size)
	tracks := make([]Track, numTracks)
	for i, e := range ((*[1 << 30]*C.struct_track)(unsafe.Pointer(cTracks.tracks)))[:numTracks:numTracks] {
		var err error
		tracks[i], err = cTrackToTrack(e)
		if err != nil {
			return tracks[:i], err
		}
	}

	return tracks, nil
}

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

func cTrackToTrack(track *C.struct_track) (Track, error) {
	if track == nil {
		return Track{}, errors.New("cannot get info from null track")
	}

	filename := C.GoString(track.filename)

	mtime, err := getMTime(filename)
	if err != nil {
		return Track{}, err
	}

	t := Track{
		Artist:      C.GoString(track.artist),
		Album:       C.GoString(track.album),
		Genre:       C.GoString(track.genre),
		Title:       C.GoString(track.title),
		Filename:    filename,
		Composer:    C.GoString(track.composer),
		Comment:     C.GoString(track.comment),
		AlbumArtist: C.GoString(track.albumArtist),
		Grouping:    C.GoString(track.grouping),
		Year:        uint32(track.year),
		Disc:        uint32(track.disc),
		Track:       uint32(track.track),
		Bitrate:     uint32(track.bitrate),
		Length:      uint32(track.length),
		Mtime:       mtime,
	}

	if t.AlbumArtist == "" || t.AlbumArtist == "<Untagged>" {
		t.AlbumArtist = t.Artist
	}

	if t.Grouping == "" || t.Grouping == "<Untagged>" {
		t.Grouping = t.Title
	}

	return t, nil
}

func (t *Track) String() string {
	b, err := json.MarshalIndent(*t, "", "    ")
	if err != nil {
		panic(*t)
	}
	return string(b)
}

func getMTime(filename string) (uint32, error) {
	file, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	mtime := file.ModTime()

	return uint32(mtime.Unix()), nil
}
