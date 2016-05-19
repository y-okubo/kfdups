package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/y-okubo/gogfapi/gfapi"
)

func main() {

	// GlusterFS
	vol := new(gfapi.Volume)
	if vol == nil {
		log.Fatal("Can't create volume")
	}

	ret := vol.Init("server", "gfs_volume")
	if ret != 0 {
		log.Fatal("Failed to initialize volume")
	}

	vol.Mount()

	// The web service
	router := gin.Default()

	router.DELETE("/*path", func(c *gin.Context) {
		// Pramを処理する
		path := c.Param("path")

		log.Printf("DELETE %s", path)

		err := vol.Unlink(path)
		if err != nil {
			message := "Failed to remove file " + path
			log.Print(message)
			c.JSON(400, message)
			return
		}

		c.JSON(200, "OK")
		return
	})

	router.POST("/*path", func(c *gin.Context) {
		// Pramを処理する
		path := c.Param("path")

		uploadFile, _, err := c.Request.FormFile("upload")
		if err != nil {
			log.Fatal(err)
		}
		// filename := header.Filename
		// log.Println(header.Filename)
		// log.Println(header)

		data, err := ioutil.ReadAll(uploadFile)
		if err != nil {
			message := "Failed to read upload file " + path
			log.Print(message)
			c.JSON(400, message)
			return
		}

		log.Print(data)

		file, err := vol.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)
		if err != nil {
			// c.JSON(400, err)
			message := "Failed to open path " + path
			log.Print(message)
			c.JSON(400, message)
			return
		}

		defer file.Close()

		n, err := file.Write(data)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Write: %d", n)

		c.JSON(200, "OK")
		return
	})

	router.GET("/*path", func(c *gin.Context) {
		// Pramを処理する
		path := c.Param("path")

		file, err := vol.Open(path)
		if err != nil {
			// c.JSON(400, err)
			message := "Failed to open path " + path
			log.Print(message)
			c.JSON(400, message)
			return
		}
		log.Print(file)

		stat, err := file.Stat()
		if err != nil {
			// c.JSON(400, err)
			message := "Failed to read stat " + path
			log.Print(message)
			c.JSON(400, message)
			return
		}

		if stat.IsDir() {
			entries, err := file.Readdir(100)
			if err != nil {
				// c.JSON(400, err)
				message := "Failed to read directory " + path
				log.Print(message)
				c.JSON(400, message)
				return
			}

			r := []map[string]string{}
			for _, entry := range entries {
				// log.Printf("%s\t%d\t%s\t%s", entry.Name(), entry.Size(), entry.IsDir(), entry.ModTime())
				r = append(r, map[string]string{
					"Name":    entry.Name(),
					"Size":    strconv.FormatInt(entry.Size(), 10),
					"ModTime": entry.ModTime().Format("2006-01-02 11:11:11.00"),
					"IsDir":   strconv.FormatBool(entry.IsDir()),
				})
			}

			c.JSON(200, r)
		} else {
			// サイズを取得
			size := stat.Size()
			b := make([]byte, size+1, size+1)

			len, err := file.Read(b)
			if err != nil {
				// c.JSON(400, err)
				message := "Failed to read file " + path
				log.Print(message)
				c.JSON(400, message)
				return
			}
			log.Print(len)
			c.Data(200, "application/octet-stream", b)
		}

	})

	router.Run(":8080")
}
