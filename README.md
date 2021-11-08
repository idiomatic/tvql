# tvql

A GraphQL video library server (and first schema /flex).

See [schema](graph/schema.graphqls).

* demonstrates query filters
* demonstrates pagination cursors (`video(paginate: {first:3, after:$id}) ...`)
* walks a filesystem tree
* extracts iTunes mp4 metadata (`video { title description genre }`)
* resizes extracted cover art on-the-fly (`artwork(geometry: {height: 360}) ...`)
* coalesces similarly titled videos from different transcode profiles (or "renditions") (`video { renditions { url size }}`)
* generates URL to served video HTTP resource


## miscellaneous mp4 specs

http://atomicparsley.sourceforge.net/mpeg-4files.html

https://mutagen.readthedocs.io/en/latest/api/mp4.html
