# tvql

A GraphQL video library server (and first GraphQL schema /flex).

See [schema](graph/schema.graphqls).

### query filters

	video(id: $id) ...

	videos(title: $title) ...

### query pagination cursors

	videos(paginate: { first: 3, after: $id }) ...

### filesystem tree traversal

Finds .m4v files within `$ROOT`.

### resource URL generation and embedded HTTP server

http://localhost:$PORT/video/...

### video rendition coalescing

Combining similar videos from different transcode profiles.

    videos { title releaseYear renditions { url size quality { resolution videoCodec transcodeBudget } } }

### iTunes "mp4" metadata extraction

	videos { title description genre artwork }

### on-the-fly cover art resizing

    videos { artwork(geometry: { height: 360 }) }


## miscellaneous mp4 specs

http://atomicparsley.sourceforge.net/mpeg-4files.html

https://mutagen.readthedocs.io/en/latest/api/mp4.html
