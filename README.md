# tvql

A GraphQL video library server (and first GraphQL schema /flex).

See [schema](graph/schema.graphqls).

This demonstrates:

* query filters (_e.g._, `video(id: $id) ...` or `videos(title: $title) ...`)
* query pagination cursors (_e.g._, `videos(paginate: {first:3, after:$id}) ...`)
* iTunes "mp4" metadata extraction (_e.g._, `videos { title description genre artwork }`)
* on-the-fly cover art resizing (_e.g._, `videos { artwork(geometry: {height: 360}) }`)
* video rendition coalescing: combining similar videos from different transcode profiles (_e.g._, `videos { title releaseYear renditions { url size quality { resolution videoCodec transcodeBudget }}}`)
* URL resource generation to embedded HTTP server (at `http://localhost:$PORT/video/...`)
* filesystem tree traversal (at `$ROOT`)


## miscellaneous mp4 specs

http://atomicparsley.sourceforge.net/mpeg-4files.html

https://mutagen.readthedocs.io/en/latest/api/mp4.html
