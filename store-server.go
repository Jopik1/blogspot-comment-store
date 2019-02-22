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
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

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

	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)

	router.GET("/getStatistics", func(c *gin.Context) {
		result := fmt.Sprintf(`{"Periods":%s}`, "")
		c.String(http.StatusOK, result)
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
		tmp, _ := strconv.Atoi(batchID)
		batchID = strconv.FormatInt(int64(tmp), 10) // make sure the value is strictly numeric to avoid path injection
		tmp, _ = strconv.Atoi(batchKey)
		batchKey = strconv.FormatInt(int64(tmp), 10)

		firstPart := batchID[:1]
		fileName := batchID + "." + batchKey + ".json.gz"
		targetPath := "./uploaded/" + firstPart + "/" + fileName

		ctx.Header("Content-Description", "File Transfer")
		ctx.Header("Content-Transfer-Encoding", "binary")
		ctx.Header("Content-Disposition", "attachment; filename="+fileName)
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.File(targetPath)
	})

	router.POST("/submitBatchUnit", func(c *gin.Context) {
		fmt.Println("POST")
		batchID := c.PostForm("batchID")
		batchKey := c.PostForm("batchKey")
		workerID := c.PostForm("workerID")
		version := c.PostForm("version")
		remoteIP := c.Request.RemoteAddr

		logLine := fmt.Sprintf(`submitBatchWorkUnit batchID:%s batchKey:%s workerID:%s version:%s IP:%s`, batchID, batchKey, workerID, version, remoteIP)

		fmt.Println(logLine)

		file, err := c.FormFile("data")
		newWorkBatch := workBatch{}
		newWorkBatch.Size = -1
		if err != nil {
			newWorkBatch.Message = fmt.Sprintf(`Error recieving form file: %s`, err.Error())
			c.JSON(http.StatusBadRequest, newWorkBatch)
			return
		}

		firstPart := batchID[:1]
		path := "./uploaded/" + firstPart + "/"
		genPath(&path)
		fmt.Println(path)
		filename := path + batchID + "." + batchKey + ".json.gz"
		if err := c.SaveUploadedFile(file, filename); err != nil {
			newWorkBatch.Message = fmt.Sprintf(`Error recieving file: %s`, err.Error())
			c.JSON(http.StatusBadRequest, newWorkBatch)
			return
		}
		if false {
			fmt.Println(randGen.Int63())
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