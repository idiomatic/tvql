package mp4

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"

	mp4 "github.com/abema/go-mp4"
)

type File struct {
	file io.ReadSeeker
	ilst *mp4.BoxInfo
	covr mp4.BoxInfo
	desc *mp4.Data
	gen  *mp4.Data
	too  *mp4.Data
	stik *mp4.Data
	sonm *mp4.Data
	nam  *mp4.Data
	tvnn *mp4.Data
	tvsh *mp4.Data
	sosn *mp4.Data
	tvsn *mp4.Data
	tves *mp4.Data
	tven *mp4.Data
	day  *mp4.Data
	hdvd *mp4.Data
}

func NewFile(file io.ReadSeeker) *File {
	return &File{file: file, ilst: nil}
}

func (f *File) survey() error {
	if f.ilst != nil {
		return nil
	}

	// ExtractBoxWithPayload(r, ilst, ...) is broken when descending Ilst atomss.
	// Walk all atoms instead, and reconstruct parentage.

	ilstBoxes, err := mp4.ExtractBox(f.file, nil, mp4.BoxPath{
		mp4.BoxTypeMoov(), mp4.BoxTypeUdta(), mp4.BoxTypeMeta(), mp4.BoxTypeIlst(),
	})
	if err != nil {
		return err
	}
	f.ilst = ilstBoxes[0]

	_, err = mp4.ReadBoxStructureFromInternal(f.file, f.ilst, func(handle *mp4.ReadHandle) (interface{}, error) {
		boxInfo := handle.BoxInfo

		if !boxInfo.IsSupportedType() {
			return nil, nil
		}

		if boxInfo.Context.UnderIlstMeta && len(handle.Path) > 1 {
			parent := handle.Path[len(handle.Path)-2]

			if parent == mp4.StrToBoxType("covr") {
				f.covr = boxInfo
			} else {
				box, _, err := handle.ReadPayload()
				if err != nil {
					return nil, err
				}
				if data, ok := box.(*mp4.Data); ok {
					switch parent {
					case mp4.StrToBoxType("desc"):
						f.desc = data
					case mp4.BoxType{0xA9, 'g', 'e', 'n'}:
						f.gen = data
					case mp4.BoxType{0xA9, 't', 'o', 'o'}:
						f.too = data
					case mp4.StrToBoxType("stik"):
						f.stik = data
					case mp4.StrToBoxType("sonm"):
						f.sonm = data
					case mp4.BoxType{0xA9, 'n', 'a', 'm'}:
						f.nam = data
					case mp4.StrToBoxType("tvnn"):
						// tv network
					case mp4.StrToBoxType("tvsh"):
						f.tvsh = data
					case mp4.StrToBoxType("tvsn"):
						f.tvsn = data
					case mp4.StrToBoxType("tves"):
						f.tves = data
					case mp4.StrToBoxType("tven"):
						// tv episode id
					case mp4.BoxType{0xA9, 'd', 'a', 'y'}:
						f.day = data
					case mp4.StrToBoxType("disk"):
						// disc/discs
					case mp4.StrToBoxType("trkn"):
						// track/tracks
					case mp4.StrToBoxType("hdvd"):
						f.hdvd = data
					case mp4.StrToBoxType("cnID"), mp4.StrToBoxType("sfID"), mp4.StrToBoxType("atID"), mp4.StrToBoxType("plID"):
						// some iTunes Store id
					case mp4.StrToBoxType("aART"):
						// depends on kind
					case mp4.BoxType{0xA9, 'A', 'R', 'T'}:
						// depends on kind
					case mp4.BoxType{0xA9, 'a', 'l', 'b'}:
						// depends on kind
					case mp4.StrToBoxType("----"):
						// free
					}
				}
			}
		}

		return handle.Expand()
	})

	return nil
}

func (f *File) readBox(bi mp4.BoxInfo) (mp4.IBox, error) {
	if _, err := bi.SeekToPayload(f.file); err != nil {
		return nil, err
	}

	box, _, err := mp4.UnmarshalAny(f.file, bi.Type, bi.Size-bi.HeaderSize, bi.Context)
	if err != nil {
		return nil, err
	}

	return box, nil
}

func (f *File) readData(bi mp4.BoxInfo) (*mp4.Data, error) {
	box, err := f.readBox(bi)
	if err != nil {
		return nil, err
	}

	data, ok := box.(*mp4.Data)
	if !ok {
		errors.New("coercion error")
	}

	return data, nil
}

func (f *File) Title() (string, error) {
	if err := f.survey(); err != nil {
		return "", err
	}

	if f.nam == nil {
		return "", errors.New("(c)nam atom missing")
	}

	return string(f.nam.Data), nil
}

func (f *File) SortTitle() (string, error) {
	if err := f.survey(); err != nil {
		return "", err
	}

	if f.sonm == nil {
		return "", errors.New("sonm atom missing")
	}

	return string(f.sonm.Data), nil
}

func (f *File) ReleaseDate() (string, error) {
	if err := f.survey(); err != nil {
		return "", err
	}

	if f.day == nil {
		return "", errors.New("(c)day atom missing")
	}

	return string(f.day.Data), nil
}

func (f *File) Genre() (string, error) {
	if err := f.survey(); err != nil {
		return "", err
	}

	if f.gen == nil {
		return "", errors.New("(c)gen atom missing")
	}

	return string(f.gen.Data), nil
}

func (f *File) Description() (string, error) {
	if err := f.survey(); err != nil {
		return "", err
	}

	if f.desc == nil {
		return "", errors.New("desc atom missing")
	}

	return string(f.desc.Data), nil
}

func (f *File) CoverArt() ([]byte, error) {
	if err := f.survey(); err != nil {
		return nil, err
	}

	data, err := f.readData(f.covr)
	if err != nil {
		return nil, err
	}

	return data.Data, nil
}

func (f *File) MediaKind() (string, error) {
	if err := f.survey(); err != nil {
		return "", err
	}

	if f.stik == nil {
		return "", errors.New("stik atom missing")
	}

	switch f.stik.Data[0] {
	case 9:
		return "Movie", nil
	case 10:
		return "TV Show", nil
	}

	return "", fmt.Errorf("unknown stik %d\n", f.stik.Data[0])
}

func (f *File) TVShowName() (string, error) {
	if err := f.survey(); err != nil {
		return "", err
	}

	if f.tvsh == nil {
		return "", errors.New("tvsh atom missing")
	}

	return string(f.tvsh.Data), nil
}

func (f *File) TVSortShowName() (string, error) {
	if err := f.survey(); err != nil {
		return "", err
	}

	if f.sosn == nil {
		return "", errors.New("sosn atom missing")
	}

	return string(f.sosn.Data), nil
}

func (f *File) TVSeason() (int, error) {
	if err := f.survey(); err != nil {
		return 0, err
	}

	if f.tvsn == nil {
		return 0, errors.New("tvsn atom missing")
	}

	var season int32
	binary.Read(bytes.NewBuffer(f.tvsn.Data), binary.BigEndian, &season)

	return int(season), nil
}

func (f *File) TVEpisode() (int, error) {
	if err := f.survey(); err != nil {
		return 0, err
	}

	if f.tvsn == nil {
		return 0, errors.New("tves atom missing")
	}

	var episode int32
	binary.Read(bytes.NewBuffer(f.tves.Data), binary.BigEndian, &episode)

	return int(episode), nil
}
