# tvql

A GraphQL video library server (and first GraphQL schema /flex).

See [schema](graph/schema.graphqls).

### query filters

    query ById($id: ID!) { video(id: $id) { ... } }

    query ByTitle($title: String!) { videos(title: $title) { ... } }

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

### video rendition coalescing

Similar videos from different transcode profiles are combined.

    query {
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

### iTunes metadata extraction

    query {
      videos {
        title
        releaseYear
        description
        genre
        #artwork
        episode {
          series {
            name
          }
          season
          episode
      }
    }

### extracted or computed sortable names

    query {
      videos {
        sortTitle
        episode {
          series {
            sortName
          }
        }
      }
    }

### sorted results
    
    query {
      videos {
        sortTitle
      }
    }

    query {
      series {
        sortName
      }
    }

    query {
      episodes {
        series {
          sortName
        }
        season
        episode
      }
    }

### on-the-fly cover art resizing

    query {
      videos {
        artwork(geometry: { height: 72 })
      }
    }


## miscellaneous mp4 specs

http://atomicparsley.sourceforge.net/mpeg-4files.html

https://mutagen.readthedocs.io/en/latest/api/mp4.html
