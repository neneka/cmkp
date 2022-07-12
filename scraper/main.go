package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"log"
	"net/url"
	"strconv"
	"time"
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("MYSQL_USERNAME", "root")
	viper.SetDefault("MYSQL_PASSWORD", "password")
	viper.SetDefault("MYSQL_HOSTNAME", "localhost")
	viper.SetDefault("MYSQL_DATABASE", "cmkp")
	viper.SetDefault("MYSQL_PORT", "3306")
}

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true",
		viper.GetString("MYSQL_USERNAME"),
		viper.GetString("MYSQL_PASSWORD"),
		viper.GetString("MYSQL_HOSTNAME"),
		viper.GetString("MYSQL_PORT"),
		viper.GetString("MYSQL_DATABASE"),
	))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	client := makeClient()
	if err := client.login(viper.GetString("WEBCATALOG_USERNAME"), viper.GetString("WEBCATALOG_PASSWORD")); err != nil {
		log.Fatal(err)
	}

	res, err := client.getBoothSyllabary()
	if err != nil {
		log.Fatal(err)
	}
	booths, err := parseBoothSyllabary(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	res.Body.Close()

	for k, v := range booths {
		res, err := client.getBooth(v)
		if err != nil {
			log.Fatal(err)
		}
		booth, err := parseBooth(res.Body)
		if err != nil {
			log.Fatal(err)
		}
		res.Body.Close()

		dbStruct := booth.convertToDBStruct()
		db.Exec("INSERT INTO circles(name, author, hall, day, block, space, location_type, genre, website, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())",
			dbStruct.Name,
			dbStruct.Author,
			dbStruct.Hall,
			dbStruct.Day,
			dbStruct.Block,
			dbStruct.Space,
			dbStruct.LocationType,
			dbStruct.Genre,
			dbStruct.Website)
		log.Println(k)
		time.Sleep(time.Second / 2)
	}

	for _, day := range []int{1, 2, 3, 4} {
		for page := 1; ; page++ {
			q := url.Values{}
			q.Set("day", strconv.Itoa(day))
			q.Set("page", strconv.Itoa(page))
			res, err := client.getCircleList(q)
			if err != nil {
				log.Fatal(err)
			}
			model, err := extractModel(res.Body)
			if err != nil {
				log.Fatal(err)
			}
			res.Body.Close()
			circles, err := parseCircleList(model)
			if err != nil {
				log.Fatal(err)
			}
			if len(circles) == 0 {
				break
			}

			for _, v := range circles {
				dbStruct := v.convertToDBStruct()
				db.Exec("INSERT INTO circles(catalog_id, name, author, hall, day, block, space, location_type, genre, website, pixiv_id, twitter_id, niconico_id, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW())",
					dbStruct.CatalogID,
					dbStruct.Name,
					dbStruct.Author,
					dbStruct.Hall,
					dbStruct.Day,
					dbStruct.Block,
					dbStruct.Space,
					dbStruct.LocationType,
					dbStruct.Genre,
					dbStruct.Website,
					dbStruct.PixivID,
					dbStruct.TwitterID,
					dbStruct.NiconicoID,
				)
			}

			log.Println(fmt.Sprintf("%d日目%dページ", day, page))
			time.Sleep(time.Second)
		}
	}
}
