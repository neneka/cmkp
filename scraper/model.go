package main

import (
	"gopkg.in/guregu/null.v3"
	"strconv"
	"strings"
)

type DBCircle struct {
	CatalogID    null.Int    `db:"catalog_id"`
	Name         string      `db:"name"`
	Author       string      `db:"author"`
	Hall         string      `db:"hall"`
	Day          int         `db:"day"`
	Block        string      `db:"block"`
	Space        string      `db:"space"`
	LocationType int         `db:"location_type"`
	Genre        string      `db:"genre"`
	PixivID      null.Int    `db:"pixiv_id"`
	TwitterID    null.String `db:"twitter_id"`
	NiconicoID   null.Int    `db:"niconico_id"`
	Website      string      `db:"website"`
}

type Booth struct {
	Name        string
	No          int
	Summery     string
	Description string
	Links       []string
}

func (b *Booth) convertToDBStruct() *DBCircle {
	s := &DBCircle{
		Name:         b.Name,
		Author:       b.Name,
		Hall:         "南",
		Day:          0,
		Block:        "",
		Space:        strconv.Itoa(b.No),
		LocationType: 0,
		Genre:        "企業",
		Website:      "",
	}
	if len(b.Links) > 0 {
		s.Website = b.Links[0]
	}
	return s
}

type Circle struct {
	Id          int         `json:"Id"`
	CircleId    int         `json:"CircleId"`
	Name        string      `json:"Name"`
	Author      string      `json:"Author"`
	Hall        string      `json:"Hall"`
	Day         string      `json:"Day"`
	Block       string      `json:"Block"`
	Space       string      `json:"Space"`
	Loc         int         `json:"Loc"`
	Genre       string      `json:"Genre"`
	PixivUrl    null.String `json:"PixivUrl"`
	TwitterUrl  null.String `json:"TwitterUrl"`
	NiconicoUrl null.String `json:"NiconicoUrl"`
	WebSite     string      `json:"WebSite"`
	Description string      `json:"Description"`
}

func (c *Circle) convertToDBStruct() *DBCircle {
	return &DBCircle{
		CatalogID:    null.IntFrom(int64(c.Id)),
		Name:         c.Name,
		Author:       c.Author,
		Hall:         c.Hall,
		Day:          toDay(c.Day),
		Block:        c.Block,
		Space:        c.Space,
		LocationType: getLocationType(c),
		Genre:        c.Genre,
		PixivID:      cutPixivID(c.PixivUrl),
		TwitterID:    cutTwitterID(c.TwitterUrl),
		NiconicoID:   cutNiconicoID(c.NiconicoUrl),
		Website:      c.WebSite,
	}
}

func toDay(s string) int {
	switch s {
	case "土":
		return 1
	case "日":
		return 2
	default:
		panic("unknown day")
	}
}

func cutPixivID(url null.String) null.Int {
	if !url.Valid {
		return null.Int{}
	}
	id, err := strconv.Atoi(strings.TrimPrefix(url.String, "http://www.pixiv.net/member.php?id="))
	if err != nil {
		return null.Int{}
	}
	return null.IntFrom(int64(id))
}

func cutNiconicoID(url null.String) null.Int {
	if !url.Valid {
		return null.Int{}
	}
	id, err := strconv.Atoi(strings.TrimPrefix(url.String, "http://www.nicovideo.jp/user/"))
	if err != nil {
		return null.Int{}
	}
	return null.IntFrom(int64(id))
}

func cutTwitterID(url null.String) null.String {
	if !url.Valid {
		return null.String{}
	}
	id := strings.TrimPrefix(url.String, "https://twitter.com/")
	if len(id) == 0 {
		return null.String{}
	}
	return null.StringFrom(id)
}

func getLocationType(c *Circle) int {
	num, _ := strconv.Atoi(strings.TrimRight(c.Space, "ab"))
	switch c.Hall {
	case "東":
		switch c.Block {
		case "A":
			shutters := []int{4,5,6,15,16,17,28,29,30,45,46,47,75,76,77,86,87,88}
			for _, v := range shutters {
				if v == num {
					return 2
				}
			}
			return 1
		case "シ":
			shutters := []int{4,5,6,15,16,17,28,29,30,45,46,47,62,63,64,75,76,77,86,87,88}
			for _, v := range shutters {
				if v == num {
					return 2
				}
			}
			return 1
		default:
			return 0
		}
	case "西":
		switch c.Block {
		case "あ":
			shutters := []int{33,34,41,42,49,50}
			for _, v := range shutters {
				if v == num {
					return 2
				}
			}
			return 1
		case "め":
			shutters := []int{19,20,33,34,41,42,49,50}
			for _, v := range shutters {
				if v == num {
					return 2
				}
			}
			return 1
		default:
			return 0
		}
	default:
		return 0
	}
}
