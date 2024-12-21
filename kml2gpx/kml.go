package kml2gpx

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/twpayne/go-gpx"
	_ "github.com/twpayne/go-gpx"
	_ "github.com/twpayne/go-kml"
)

type kml struct {
	Document document `xml:"Document"`
}

type document struct {
	Name       string      `xml:"name"`
	StyleMaps  []styleMap  `xml:"StyleMap"`
	PlaceMarks []placeMark `xml:"Placemark"`
}

type placeMark struct {
	Name        string       `xml:"name"`
	LineStrings []lineString `xml:"LineString"`
}

type lineString struct {
	Coordinates string `xml:"coordinates"`
}

type styleMap struct {
	ID string `xml:"id,attr"`
}

func Convert(data []byte) (gpxdata *gpx.GPX, err error) {
	k := &kml{}
	g := &gpx.GPX{
		Metadata: &gpx.MetadataType{},
	}
	err = xml.NewDecoder(bytes.NewBuffer(data)).Decode(k)
	if err != nil {
		return nil, fmt.Errorf("error decoding data: %w", err)
	}
	g.Metadata.Name = k.Document.Name
	for _, p := range k.Document.PlaceMarks {
		trk := &gpx.TrkType{
			Name: p.Name,
		}
		if len(p.LineStrings) == 0 {
			continue
		}
		trkSeg := &gpx.TrkSegType{}
		for _, c := range strings.Split(p.LineStrings[0].Coordinates, "\n") {
			if lon, lat := splitCoordinates(c); lon != 0 && lat != 0 {
				trkSeg.TrkPt = append(trkSeg.TrkPt, &gpx.WptType{Lat: lat, Lon: lon})
			}
		}
		trk.TrkSeg = append(trk.TrkSeg, trkSeg)
		g.Trk = append(g.Trk, trk)
	}
	return g, nil
}

func splitCoordinates(in string) (lon float64, lat float64) {
	if strings.TrimSpace(in) == "" {
		return 0, 0
	}
	split := strings.Split(in, ",")
	for i := range split {
		switch i {
		case 0:
			lon, _ = strconv.ParseFloat(strings.TrimSpace(split[i]), 64)
		case 1:
			lat, _ = strconv.ParseFloat(strings.TrimSpace(split[i]), 64)
		}
	}
	return lon, lat
}
