/*
* Server for recieving work units produced
* by tech234a and afrmtbl blogspot-comment-backup project
* This program is free software: you can redistribute it and/or modify
* it under the terms of the GNU General Public License as published by
* the Free Software Foundation, version 3.
*
* This program is distributed in the hope that it will be useful, but
* WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
* General Public License for more details.
*
* You should have received a copy of the GNU General Public License
* along with this program. If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"runtime"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

var storeDirectory = "./uploaded/"
var regLeaveOnlyDigits = regexp.MustCompile("[^0-9]+")

type workBatch struct {
	BatchID string `json:"batch_id"`
	Message string `json:"message,omitempty"`
	Size    int64  `json:"size"`
}

func emptyPrintf(format string, v ...interface{}) {
	return
}

// JSONMarshalIndentNoEscapeHTML allow proper json formatting
func JSONMarshalIndentNoEscapeHTML(i interface{}, prefix string, indent string) ([]byte, error) {
	buf := &bytes.Buffer{}
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent(prefix, indent)

	err := encoder.Encode(i)
	return buf.Bytes(), err
}

func genPath(path *string) {
	// create directory if it doesnt exist
	if _, err := os.Stat(*path); err != nil { // os.IsNotExist(err)
		err = os.MkdirAll(*path, 0755)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			runtime.Goexit()
		}
	}

}

func main() {
	runtime.GOMAXPROCS(250)
	//gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	// Set a memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 2048 << 20 // 2048 MiB

	router.StaticFS("/uploadedBatches", gin.Dir(storeDirectory, true))

	router.GET("/getBatchUnit", func(ctx *gin.Context) {
		batchID, batchIDok := ctx.GetQuery("batchID")
		batchKey, batchKeyok := ctx.GetQuery("batchKey")
		newWorkBatch := workBatch{}

		if !batchIDok || !batchKeyok {
			newWorkBatch.Message = fmt.Sprintf(`batchID or batchKey parameters missing`)
			ctx.JSON(http.StatusBadRequest, newWorkBatch)
			return
		}
		batchID = regLeaveOnlyDigits.ReplaceAllString(batchID, "")
		batchKey = regLeaveOnlyDigits.ReplaceAllString(batchKey, "")

		firstPart := batchID[:1]
		fileName := batchID + "." + batchKey + ".json.gz"
		targetPath := storeDirectory + firstPart + "/" + fileName

		_, err := os.Stat(targetPath)
		newWorkBatch.BatchID = batchID
		if err == nil {
			ctx.Header("Content-Description", "File Transfer")
			ctx.Header("Content-Transfer-Encoding", "binary")
			ctx.Header("Content-Disposition", "attachment; filename="+fileName)
			ctx.Header("Content-Type", "application/octet-stream")
			ctx.File(targetPath)
		} else {
			newWorkBatch.Message = "Batch doesn't exist " + err.Error()
			newWorkBatch.Size = -1
			ctx.JSON(http.StatusNotFound, newWorkBatch)
		}
		return
	})

	router.GET("/getVerifyBatchUnit", func(ctx *gin.Context) {
		batchID, batchIDok := ctx.GetQuery("batchID")
		batchKey, batchKeyok := ctx.GetQuery("batchKey")
		newWorkBatch := workBatch{}

		if !batchIDok || !batchKeyok {
			newWorkBatch.Message = fmt.Sprintf(`batchID or batchKey parameters missing`)
			ctx.JSON(http.StatusBadRequest, newWorkBatch)
			return
		}
		batchID = regLeaveOnlyDigits.ReplaceAllString(batchID, "")
		batchKey = regLeaveOnlyDigits.ReplaceAllString(batchKey, "")

		firstPart := batchID[:1]
		fileName := batchID + "." + batchKey + ".json.gz"
		targetPath := storeDirectory + firstPart + "/" + fileName

		fileInfo, err := os.Stat(targetPath)
		newWorkBatch.BatchID = batchID
		if err == nil {
			newWorkBatch.Message = "Batch Exists"
			newWorkBatch.Size = fileInfo.Size()
			ctx.JSON(http.StatusOK, newWorkBatch)
			return
		} else {
			newWorkBatch.Message = "Batch doesn't exist"
			newWorkBatch.Size = -1
			ctx.JSON(http.StatusNotFound, newWorkBatch)
			return
		}
	})

	router.POST("/submitBatchUnit", func(c *gin.Context) {
		fmt.Println("POST")
		batchID := regLeaveOnlyDigits.ReplaceAllString(c.PostForm("batchID"), "")
		batchKey := regLeaveOnlyDigits.ReplaceAllString(c.PostForm("batchKey"), "")
		workerID := c.PostForm("workerID")
		version := c.PostForm("version")

		remoteIP := c.Request.RemoteAddr

		logLine := fmt.Sprintf(`submitBatchUnit batchID:%s batchKey:%s workerID:%s version:%s IP:%s`, batchID, batchKey, workerID, version, remoteIP)

		fmt.Println(logLine)

		file, err := c.FormFile("data")
		newWorkBatch := workBatch{}
		newWorkBatch.Size = -1
		newWorkBatch.BatchID = batchID
		if err != nil {

			newWorkBatch.Message = fmt.Sprintf(`Error recieving form file: %s`, err.Error())
			c.JSON(http.StatusBadRequest, newWorkBatch)
			return
		}

		firstPart := batchID[:1]
		path := storeDirectory + firstPart + "/"
		genPath(&path)
		fmt.Println(path)
		filename := path + batchID + "." + batchKey + ".json.gz"
		if err := c.SaveUploadedFile(file, filename); err != nil {
			newWorkBatch.Message = fmt.Sprintf(`Error recieving file: %s`, err.Error())
			c.JSON(http.StatusBadRequest, newWorkBatch)
			return
		}

		newWorkBatch.Message = "Batch Accepted"
		fileInfo, err := os.Stat(filename)
		if err == nil {
			newWorkBatch.Size = fileInfo.Size()
		} else {
			fmt.Println("Stat error", err)
		}

		c.JSON(http.StatusOK, newWorkBatch)
	})

	router.Run(":80")

}
