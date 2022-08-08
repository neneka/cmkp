package model

import (
	"context"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

func getCircleLoader(ctx context.Context) (*CircleLoader, bool) {
	v, e := ctx.Value(LoadersKey).(*Loaders)
	return v.Circle, e
}

func createCircleLoader(ctx context.Context) *CircleLoader {
	db := orm(ctx)

	return &CircleLoader{
		wait:     1 * time.Millisecond,
		maxBatch: 1000,
		fetch: func(keys []int) (items []*Circle, errors []error) {
			items = make([]*Circle, len(keys))
			errors = make([]error, len(keys))

			var tmp []*Circle
			if err := db.Where("id IN (?)", keys).Find(&tmp).Error; err != nil {
				panic(err)
			}

			// now, tmp is sorted by id, not keys

			for i, key := range keys {
				ok := false
				for _, v := range tmp {
					if v.ID == key {
						ok = true
						items[i] = v
						break
					}
				}
				if !ok {
					errors[i] = gorm.ErrRecordNotFound
				}
			}
			return
		},
	}
}

type Circle struct {
	ID           int       `gorm:"primary_key"      json:"id"`
	CatalogID    *int      `                        json:"-"`
	Name         string    `sql:"type:varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL" json:"name"`
	Author       string    `sql:"type:varchar(30) COLLATE utf8mb4_unicode_ci NOT NULL" json:"author"`
	Hall         string    `gorm:"size:1"           json:"hall"`
	Day          int       `                        json:"day"`
	Block        string    `gorm:"size:1"           json:"block"`
	Space        string    `                        json:"space"`
	LocationType int       `                        json:"locationType"`
	Genre        string    `                        json:"genre"`
	PixivID      *int      `                        json:"pixivId"`
	TwitterID    *string   `gorm:"type:varchar(15)" json:"twitterId"`
	NiconicoID   *int      `                        json:"niconicoId"`
	Website      string    `gorm:"type:text"        json:"website"`
	UpdatedAt    time.Time `gorm:"precision:6"      json:"updatedAt"`
}

func (c *Circle) LocationString(ctx context.Context, day bool) (string, error) {
	if c.Day == 0 {
		if day {
			return "企業" + c.Hall + c.Space, nil
		} else {
			return c.Hall + c.Space, nil
		}
	} else {
		if day {
			return strconv.Itoa(c.Day) + "日目" + c.Hall + c.Block + c.Space, nil
		} else {
			return c.Hall + c.Block + c.Space, nil
		}
	}
}

func (c *Circle) Memos(ctx context.Context) ([]*CircleMemo, error) {
	return GetCircleMemosByCircleID(ctx, c.ID)
}

func (c *Circle) Items(ctx context.Context) ([]*Item, error) {
	return GetItemsByCircleID(ctx, c.ID)
}

func (c *Circle) RequestingUser(ctx context.Context) ([]*User, error) {
	if IsGranted(ctx, getCtxUserRole(ctx), RolePlanner) {
		return GetRequestingUserByCircleID(ctx, c.ID)
	} else {
		return nil, ErrForbidden
	}
}

func (c *Circle) RequestedItems(ctx context.Context, userId *int) ([]*Item, error) {
	if userId != nil {
		if *userId == getCtxUserId(ctx) || IsGranted(ctx, getCtxUserRole(ctx), RolePlanner) {
			return GetRequestedItemsByCircleID(ctx, c.ID, userId)
		}
	} else if IsGranted(ctx, getCtxUserRole(ctx), RolePlanner) {
		return GetRequestedItemsByCircleID(ctx, c.ID, nil)
	}
	return nil, ErrForbidden
}

func (c *Circle) Prioritized(ctx context.Context) ([]*PriorityRank, error) {
	if IsGranted(ctx, getCtxUserRole(ctx), RolePlanner) {
		return GetPriorityRankListByCircleID(ctx, c.ID)
	} else {
		return nil, ErrForbidden
	}
}

func GetCircleByID(ctx context.Context, id int) (*Circle, error) {
	if loader, ok := getCircleLoader(ctx); ok {
		return loader.Load(id)
	}

	circle := Circle{}
	if err := orm(ctx).First(&circle, id).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	return &circle, nil
}

func GetCircleByItemID(ctx context.Context, itemID int) (*Circle, error) {
	query := orm(ctx).
		Model(Circle{}).
		Where("id = ?", orm(ctx).Model(&Item{ID: itemID}).Select("circle_id").SubQuery())

	if loader, ok := getCircleLoader(ctx); ok {
		ids := make([]int, 0)
		if err := query.Pluck("circles.id", &ids).Error; err != nil {
			panic(err)
		}

		if len(ids) != 1 {
			return nil, gorm.ErrRecordNotFound
		}

		return loader.Load(ids[0])
	}

	v := Circle{}
	if err := query.First(&v).Error; err != nil {
		return nil, panicUnlessNotFound(err)
	}
	return &v, nil
}

func SearchCircles(ctx context.Context, q string, days []int) ([]*Circle, error) {
	tmp := "%" + q + "%"

	query := orm(ctx).
		Model(Circle{}).
		Limit(100).
		Where("name LIKE ? OR author LIKE ?", tmp, tmp)
	if len(days) > 0 {
		query = query.Where("day IN (?)", days)
	}

	if loader, ok := getCircleLoader(ctx); ok {
		ids := make([]int, 0)
		if err := query.Pluck("circles.id", &ids).Error; err != nil {
			panic(err)
		}

		arr, errs := loader.LoadAll(ids)
		for _, v := range errs {
			if v != nil {
				return nil, v
			}
		}
		return arr, nil
	}

	arr := make([]*Circle, 0)
	if err := query.Find(&arr).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func GetRequestedCirclesByDay(ctx context.Context, day int) ([]*Circle, error) {
	query := orm(ctx).
		Model(Circle{}).
		Joins("INNER JOIN (SELECT items.circle_id as id FROM items INNER JOIN user_request_items ON items.id = user_request_items.item_id GROUP BY items.circle_id) t ON circles.id = t.id").
		Order("circles.hall").Order("circles.block").Order("circles.space")
	if day > -1 {
		query = query.Where("circles.day = ?", day)
	}

	if loader, ok := getCircleLoader(ctx); ok {
		ids := make([]int, 0)
		if err := query.Pluck("circles.id", &ids).Error; err != nil {
			panic(err)
		}

		arr, errs := loader.LoadAll(ids)
		for _, v := range errs {
			if v != nil {
				return nil, v
			}
		}
		return arr, nil
	}

	arr := make([]*Circle, 0)
	if err := query.Find(&arr).Error; err != nil {
		panic(err)
	}
	return arr, nil
}

func GetRequestedCirclesByUser(ctx context.Context, userID int) ([]*Circle, error) {
	query := orm(ctx).
		Model(Circle{}).
		Joins("INNER JOIN (SELECT items.circle_id as id FROM items INNER JOIN user_request_items ON items.id = user_request_items.item_id WHERE user_request_items.user_id = ? GROUP BY items.circle_id) t ON circles.id = t.id", userID).
		Order("circles.day").Order("circles.hall").Order("circles.block").Order("circles.space")

	if loader, ok := getCircleLoader(ctx); ok {
		ids := make([]int, 0)
		if err := query.Pluck("circles.id", &ids).Error; err != nil {
			panic(err)
		}

		arr, errs := loader.LoadAll(ids)
		for _, v := range errs {
			if v != nil {
				return nil, v
			}
		}
		return arr, nil
	}

	arr := make([]*Circle, 0)
	if err := query.Find(&arr).Error; err != nil {
		panic(err)
	}
	return arr, nil
}
