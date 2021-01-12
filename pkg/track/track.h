#ifndef TRACK
#define TRACK

#ifdef __cplusplus
extern "C" {
#endif

#include <stdlib.h>

struct track {
    char* artist;
    char* album;
    char* genre;
    char* title;
    char* filename;
    char* composer;
    char* comment;
    char* albumArtist;
    char* grouping;
    unsigned year;
    unsigned disc;
    unsigned track;
    unsigned bitrate;
    unsigned length;
};

struct tracks {
    struct track** tracks;
    unsigned size;
};

struct tracks* getTracks(char** filenames, size_t size);
struct track* getTrack(char* filename);
void* freeTrack(struct track* track);
void* freeTracks(struct tracks* tracks);

#ifdef __cplusplus
}
#endif

#endif