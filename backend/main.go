package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/handler"
	"github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"github.com/vektah/gqlparser/gqlerror"
	"github.com/wtks/cmkp/backend/model"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"syscall"
	"time"
)

var (
	orm *gorm.DB
)

func init() {
	viper.SetDefault("INIT_ADMIN_PASSWORD", "password")
	viper.SetDefault("MYSQL_HOSTNAME", "localhost")
	viper.SetDefault("MYSQL_DATABASE", "cmkp")
	viper.SetDefault("MYSQL_USERNAME", "root")
	viper.SetDefault("MYSQL_PASSWORD", "password")
	viper.SetDefault("JWT_SECRET", "cmkpsupersecret")
	viper.SetDefault("PORT", 3000)
	viper.SetDefault("CMKP_EVENT_DAYS", 2)
	viper.AutomaticEnv()
}

func main() {
	// connect database
	var err error
	orm, err = gorm.Open("mysql", fmt.Sprintf(
		"%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=true",
		viper.GetString("MYSQL_USERNAME"),
		viper.GetString("MYSQL_PASSWORD"),
		viper.GetString("MYSQL_HOSTNAME"),
		viper.GetString("MYSQL_DATABASE"),
	))
	if err != nil {
		log.Fatal(err)
	}
	orm.DB().SetMaxOpenConns(50)

	ctx := context.WithValue(context.Background(), "orm", orm)
	// init db data
	if err := model.InitTables(ctx); err != nil {
		log.Fatal(err)
	}
	if err = initAdminUser(ctx); err != nil {
		log.Fatal(err)
	}
	if err = initDeadline(ctx); err != nil {
		log.Fatal(err)
	}

	// init web server
	e := echo.New()
	e.Validator = newValidator()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{MaxAge: 86400}))
	e.Use(middleware.Gzip())

	// routing api
	e.POST("/api/login", login)
	gql := handler.GraphQL(
		NewExecutableSchema(Config{Resolvers: &Resolver{}}),
		handler.ErrorPresenter(func(ctx context.Context, e error) *gqlerror.Error {
			if gorm.IsRecordNotFoundError(err) {
				return gqlerror.ErrorPathf(graphql.GetResolverContext(ctx).Path(), "not found")
			}
			return graphql.DefaultErrorPresenter(ctx, e)
		}),
		handler.RecoverFunc(func(ctx context.Context, err interface{}) (userMessage error) {
			fmt.Fprintln(os.Stderr, err)
			fmt.Fprintln(os.Stderr)
			debug.PrintStack()

			return errors.New("internal server error")
		}),
	)
	e.POST("/api/graphql",
		echo.WrapHandler(gql),
		middleware.JWTWithConfig(middleware.JWTConfig{
			Claims:     &jwt.StandardClaims{},
			SigningKey: []byte(viper.GetString("JWT_SECRET")),
		}),
		func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c echo.Context) error {
				claim := c.Get("user").(*jwt.Token).Claims.(*jwt.StandardClaims)
				user := &model.User{}
				if err := orm.Select("id, role").First(user, mustParseInt(claim.Subject)).Error; err != nil {
					c.Logger().Error(err)
					return echo.NewHTTPError(http.StatusInternalServerError)
				}
				c.SetRequest(c.Request().WithContext(context.WithValue(context.WithValue(context.WithValue(context.WithValue(c.Request().Context(), "userId", user.ID), "role", model.Role(user.Role)), model.ORMKey, orm), model.LoadersKey, model.CreateLoaders(ctx))))
				return next(c)
			}
		},
	)

	// start server
	go func() {
		if err := e.Start(fmt.Sprintf(":%d", viper.GetInt("PORT"))); err != nil {
			e.Logger.Info("shutting down the server")
		}
	}()

	// graceful shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err = e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func initAdminUser(ctx context.Context) error {
	u := &model.User{}
	if err := orm.Where(&model.User{Role: model.RoleAdmin}).Take(u).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			hash, _ := bcrypt.GenerateFromPassword([]byte(viper.GetString("INIT_ADMIN_PASSWORD")), bcrypt.DefaultCost)
			u = &model.User{
				Name:              "admin",
				DisplayName:       "管理人",
				EncryptedPassword: string(hash),
				Role:              model.RoleAdmin,
			}
			if err := orm.Create(u).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}

func initDeadline(ctx context.Context) error {
	for i := 0; i <= viper.GetInt("CMKP_EVENT_DAYS"); i++ {
		_, err := model.GetDeadline(ctx, i)
		if err == gorm.ErrRecordNotFound {
			if _, err := model.SetDeadline(ctx, i, time.Now().Truncate(time.Minute)); err != nil {
				return err
			}
		}
	}
	return nil
}

func login(c echo.Context) error {
	req := struct {
		User string `json:"username" validate:"required"`
		Pass string `json:"password" validate:"printascii,required,max=50"`
	}{}
	if err := bindAndValidate(c, &req); err != nil {
		return err
	}

	user := &model.User{}
	if err := orm.Where(&model.User{Name: req.User}).Take(user).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return echo.NewHTTPError(http.StatusBadRequest, "ユーザー名またはパスワードが間違っています")
		}
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if !user.CheckPassword(req.Pass) {
		return echo.NewHTTPError(http.StatusBadRequest, "ユーザー名またはパスワードが間違っています")
	}

	claims := &jwt.StandardClaims{
		Issuer:    "cmkp",
		IssuedAt:  time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 24 * 20).Unix(),
		Subject:   strconv.Itoa(user.ID),
		Audience:  user.Name,
	}

	t, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(viper.GetString("JWT_SECRET")))
	if err != nil {
		c.Logger().Error(err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return c.JSON(http.StatusOK, map[string]string{"token": t})
}
