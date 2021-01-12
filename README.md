# Dependencies

* [taglib](http://taglib.github.com/)
* [golang](https://golang.org/)

### OSX:
    brew install taglib golang

### Ubuntu:
    sudo apt-get install libtag1-dev golang

# Compile

    make

# Usage

### rbdbgen

```
Usage of bin/rbdbgen:
  -big
        use big endian database (coldfire and SH1)
  -external string
        location of music on external media
  -internal string
        location of music on internal media
  -target string
        directory to output database files to (will be created if not exists) (default "./database/")
```

### rbdbdump

```
Usage of bin/rbdbdump:
  -csv
        save as csv instead of xlsx
  -in string
        directory containing database files (default "./.rockbox/")
  -out string
        directory to output database dumps to (will be created if not exists) (default "./csv/")
```

# TODO

- [ ] Clean code
- [ ] Get tags using native go
- [ ] Add features?