#include "taglib.h"

#include <stdlib.h>
#include <taglib/fileref.h>
#include <taglib/tag.h>
#include <taglib/tpropertymap.h>
#include <string.h>

#define UNTAGGED "<Untagged>"

char* copyString(TagLib::String str) {
    bool untagged = str.size() == 0;
    if (untagged) {
        str = TagLib::String(UNTAGGED);
    }
    char* s = (char*) calloc(str.size() + 1, sizeof(char));
    if (s == NULL) return NULL;
    memcpy(s, untagged ? UNTAGGED : str.toCString(true), str.size());
    return s;
}

struct track* getTrack(char* filename) {
    TagLib::FileRef file(filename);

    if (!file.isNull()) {
        TagLib::Tag *tag = file.tag();
        TagLib::AudioProperties *properties = file.audioProperties();

        struct track* t = (struct track*) calloc(1, sizeof(struct track));
        if (t == NULL) return NULL;

        t->filename = copyString(TagLib::String(filename));
        if (t->filename == NULL) return (struct track*)  freeTrack(t);

        if (tag) {
            t->artist = copyString(tag->artist());
            if (t->artist == NULL) return (struct track*) freeTrack(t);
            t->album = copyString(tag->album());
            if (t->album == NULL) return (struct track*) freeTrack(t);
            t->genre = copyString(tag->genre());
            if (t->genre == NULL) return (struct track*) freeTrack(t);
            t->title = copyString(tag->title());
            if (t->title == NULL) return (struct track*) freeTrack(t);
            t->comment = copyString(tag->comment());
            if (t->comment == NULL) return (struct track*) freeTrack(t);
            t->year = tag->year();
            t->track = tag->track();
        }

        t->composer = copyString(TagLib::String());
        if (t->composer == NULL) return (struct track*) freeTrack(t);

        t->albumArtist = (char*) calloc(1, sizeof(char));
        if (t->albumArtist == NULL) return (struct track*) freeTrack(t);
        t->grouping = (char*) calloc(1, sizeof(char));
        if (t->grouping == NULL) return (struct track*) freeTrack(t);
        t->disc = 0;

        TagLib::PropertyMap tags = file.file()->properties();

        // composer, albumArtist, grouping, disc
        bool found[] = {false, false, false, false};
        for(TagLib::PropertyMap::ConstIterator i = tags.begin(); i != tags.end(); ++i) {
            if (found[0] && found[1] && found[2] && found[3]) break;

            TagLib::String upper = i->first.upper();
            if (upper == "COMPOSER") {
                found[0] = true;
                free(t->composer);
                t->composer = copyString(i->second.toString());
                if (t->composer == NULL) return (struct track*) freeTrack(t);
            }
            if (upper == "ALBUMARTIST" || upper == "ALBUM ARTIST" || upper == "BAND" || upper == "ENSEMBLE") {
                found[1] = true;
                free(t->albumArtist);
                t->albumArtist = copyString(i->second.toString());
                if (t->albumArtist == NULL) return (struct track*) freeTrack(t);
            }
            if (upper == "GROUPING" || upper == "ITUNES GROUPING") {
                found[2] = true;
                free(t->grouping);
                t->grouping = copyString(i->second.toString());
                if (t->grouping == NULL) return (struct track*) freeTrack(t);
            }
            if (upper == "DISCNUMBER" || upper == "DISC NUMBER") {
                found[3] = true;
                unsigned len = i->second.toString().size();
                const char* strNum = i->second.toString().toCString(true);
                for (int j = 0; j < len; j++) {
                    switch (strNum[j]) {
                        case '0': case '1': case '2': case '3': case '4':
                        case '5': case '6': case '7': case '8': case '9':
                            t->disc *= 10;
                            t->disc += strNum[j] - '0';
                            continue;
                        default:
                            j = 1;
                            len = 0;
                    }
                }
            }
        }

        if (properties) {
            t->bitrate = properties->bitrate();
            t->length = properties->lengthInMilliseconds();
        }

        return t;
    } else return NULL;
}

struct tracks* getTracks(char** filenames, size_t size) {
    auto* tracks = (struct tracks*) calloc(1, sizeof(struct tracks));

    if (tracks != nullptr) {
        tracks->tracks = (struct  track**) malloc(sizeof(struct track*) * size);
        if (tracks->tracks == NULL) {
            free(tracks);
            return NULL;
        }
        for (int i = 0; i < size; i++, tracks->size++) {
            tracks->tracks[i] = getTrack(filenames[i]);
            if (tracks->tracks[i] == NULL) {
                tracks->tracks = (struct track**) realloc(tracks->tracks, sizeof(struct track*)*size);
                if (tracks->tracks == NULL) tracks->size = 0;
                return tracks;
            }
        }
    }

    return tracks;
}

void* freeTrack(struct track* track) {
    free(track->artist);
    free(track->album);
    free(track->genre);
    free(track->title);
    free(track->composer);
    free(track->comment);
    free(track->albumArtist);
    free(track->grouping);
    free(track->filename);
    free(track);
    return nullptr;
}


void* freeTracks(struct tracks* tracks) {
    for (int i = 0; i < tracks->size; i++) {
        freeTrack(tracks->tracks[i]);
    }
    free(tracks);
    return nullptr;
}