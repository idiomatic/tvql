# tvql

A GraphQL video library server (and first GraphQL schema /flex).

See [schema](graph/schema.graphqls).

### query filters

    query ById($id: ID!) { video(id: $id) { title } }

    query ByTitle($title: String!) { videos(title: $title) { releaseYear } }

### query pagination cursors

    query SomeVids($count: Int!, $id: ID!) {
      videos(paginate: { first: $count, after: $id }) {
        ...
      }
    }

### filesystem tree traversal

Find .m4v files within `$ROOT`.

### resource URL generation and embedded HTTP server

Served at http://localhost:$PORT/video/

### video renditions

Similar videos (_i.e._, same title and release year) with different
transcode profiles are combined.

    query CombinedRenditions {
      videos {
        title
        releaseYear
        renditions {
          url
          size
          quality {
            resolution
            videoCodec
            transcodeBudget
          }
        }
      }
    }
    
    query TargetedRenditions {
      videos {
        title
        rendition {
          hd: rendition(quality: {resolution: "1080p"}) {
            url
          }
          dvd: rendition(quality: {resolution: "720p"}) {
            url
          }
          sd: rendition(quality: {resolution: "480p"}) {
            url
          }
        }
      }
    }

### iTunes metadata extraction

    query ItunesAtoms {
      videos {
        title
        releaseYear
        description
        genre
        artwork {
          url(geometry: {height: 640})
        }
        episode {
          season {
            series {
              name
            }
            season
          }
          episode
        }
      }
    }

### extracted or computed sortable titles

If video has explicit iTunes metadata for a sortable title, use that.
Otherwise, adapt the display title.

    query Sortables {
      videos {
        sortTitle
        episode {
          season {
            series {
              sortName
            }
          }
        }
      }
    }

### sorted query results
    
Compatible with pagination.

    query SortedVideos {
      videos {
        sortTitle
      }
    }

    query SortedSeries {
      series {
        sortName
      }
    }

    query SortedSeasons {
      seasons {
        series {
          sortName
        }
        season
      }
    }

    query SortedEpisodes {
      episodes {
        season {
          series {
            sortName
          }
          season
        }
        episode
      }
    }

### on-the-fly cover art resizing

    query ArtworkResizing {
      videos {
        artwork {
          base64(geometry: { height: 72 })
        }
      }
    }


## miscellaneous mp4 specs

http://atomicparsley.sourceforge.net/mpeg-4files.html

https://mutagen.readthedocs.io/en/latest/api/mp4.html
