package main

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/labstack/echo"
	"gopkg.in/gorp.v2"
)

type Controller struct {
	dbmap *gorp.DbMap
}

type Error struct {
	Error string `json:"error"`
}

type Comment struct {
	Id      int64     `json:"id" db:"id,primarykey,autoincrement"`
	Name    string    `json:"name" form:"name" db:"name,notnull,size:200"`
	Text    string    `json:"text" form:"text" validate:"required,max=20" db:"text,notnull,size:399"`
	Created time.Time `json:"created" db:"created,notnull"`
	Updated time.Time `json:"updated" db:"updated,notnull"`
}

func main() {
	dbmap := initDb()

	controller := &Controller{dbmap: dbmap}
	e := echo.New()
	e.GET("/api/comments/:id", controller.GetComment)
	e.GET("/api/comments/", controller.ListComments)
	e.POST("/api/comments/", controller.InsertComment)
	e.Static("/", "static/")
	e.Logger.Fatal(e.Start(":8989"))

}

func initDb() *gorp.DbMap {
	db, err := sql.Open("mysql", "user:password@tcp(host:port)/dbname")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.MySQLDialect{"InnoDB", "UTF8"}}
	defer dbmap.Db.Close()

	return dbmap
}

func (controller *Controller) GetComment(c echo.Context) error {
	var comment Comment
	err := controller.dbmap.SelectOne(&comment,
		"SELECT * FROM commnets WHERE id = $1", c.Param("id"))
	if err != nil {
		if err != sql.ErrNoRows {
			c.Logger().Error("SelectOne:", err)
			return c.String(http.StatusBadRequest, "Not Found")
		}
		return c.JSON(http.StatusOK, comment)
	}
	return c.JSON(http.StatusOK, comment)
}

func (controller *Controller) ListComments(c echo.Context) error {
	var comments []Comment
	_, err := controller.dbmap.Select(&comments,
		"SELECT * FROM comments ORDER BY created desc LIMIT 10")
	if err != nil {
		c.Logger().Error("Select:", err)
		return c.String(http.StatusBadRequest, "Select:"+err.Error())
	}
	return c.JSON(http.StatusOK, comments)
}

func (controller *Controller) InsertComment(c echo.Context) error {
	var comment Comment
	if err := c.Bind(&comment); err != nil {
		c.Logger().Error("Bind: ", err)
		return c.String(http.StatusBadRequest, "Bind:"+err.Error())
	}
	if err := controller.dbmap.Insert(&comment); err != nil {
		c.Logger().Error("Insert: ", err)
		return c.String(http.StatusBadRequest, "Insert: "+err.Error())
	}
	c.Logger().Infof("inserted comment: %v", comment.Id)
	return c.NoContent(http.StatusCreated)
}
