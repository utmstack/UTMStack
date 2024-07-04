package main

import (
	"bytes"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/threatwinds/go-sdk/helpers"
	"github.com/threatwinds/go-sdk/plugins"
	"google.golang.org/protobuf/encoding/protojson"
)

func Log(c *gin.Context) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		e := helpers.Logger().ErrorF(err.Error())
		e.GinError(c)
		return
	}

	var l = new(plugins.Log)
	err = protojson.Unmarshal(buf.Bytes(), l)
	if err != nil {
		e := helpers.Logger().ErrorF(err.Error())
		e.GinError(c)
		return
	}

	if l.Id == "" {
		lastId := uuid.New().String()
		l.Id = lastId
	}

	localLogsChannel <- l

	c.JSON(http.StatusOK, plugins.Ack{LastId: l.Id})
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ping": "ok"})
}

func GitHub(c *gin.Context) {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(c.Request.Body)
	if err != nil {
		e := helpers.Logger().ErrorF(err.Error())
		e.GinError(c)
		return
	}

	var l = new(plugins.Log)

	lastId := uuid.New().String()
	l.Id = lastId

	l.DataType = "github"
	l.DataSource = "github"
	l.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	l.Raw = buf.String()

	localLogsChannel <- l

	c.JSON(http.StatusOK, plugins.Ack{LastId: l.Id})
}

func (i *integration) ProcessLog(srv plugins.Integration_ProcessLogServer) error {
	for {
		l, err := srv.Recv()
		if err != nil {
			return err
		}

		if l.Id == "" {
			lastId := uuid.New().String()
			l.Id = lastId
		}

		localLogsChannel <- l

		if err := srv.Send(&plugins.Ack{LastId: l.Id}); err != nil {
			helpers.Logger().ErrorF("failed to send ack: %v", err)
			return err
		}
	}
}
