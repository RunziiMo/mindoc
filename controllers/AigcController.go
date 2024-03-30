package controllers

import (
	"strings"
	"time"

	"github.com/beego/beego/v2/client/httplib"
	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/server/web"
	"github.com/mindoc-org/mindoc/conf"
	"github.com/mindoc-org/mindoc/models"
	"github.com/mindoc-org/mindoc/utils/pagination"
)

type AigcController struct {
	BaseController
}

func (c *AigcController) ListChatMessages() {
	docid, err := c.GetInt("doc_id")
	if err != nil {
		c.JsonResult(1, err.Error())
	}
	pageIndex, _ := c.GetInt("page", 1)

	// 获取消息、分页
	messages, count, pageIndex := models.NewAigcChatMessage().QueryByDocumentId(docid, pageIndex, conf.PageSize, c.Member)
	page := pagination.PageUtil(int(count), pageIndex, conf.PageSize, messages)

	var data struct {
		DocId int             `json:"doc_id"`
		Page  pagination.Page `json:"page"`
	}
	data.DocId = docid
	data.Page = page

	c.JsonResult(0, "ok", data)
}

func (c *AigcController) DocAnalyze() {
	api := c.GetString("api")
	if api == "" {
		c.JsonResult(1, "请指定文档分析方法")
	}
	id, _ := c.GetInt("doc_id")
	logs.Notice("api %s doc_id %d", api, id)

	doc, err := models.NewDocument().Find(id)
	if err != nil {
		c.JsonResult(1, "文章不存在")
		return
	}

	m := models.NewAigcChatMessage()
	m.DocumentId = id
	if !c.isUserLoggedIn() {
		c.JsonResult(1, "请先登录，再进行交互")
		return
	}
	if len(c.Member.RealName) != 0 {
		m.Author = c.Member.RealName
	} else {
		m.Author = c.Member.Account
	}
	m.MemberId = c.Member.MemberId
	m.IPAddress = c.Ctx.Request.RemoteAddr
	m.IPAddress = strings.Split(m.IPAddress, ":")[0]
	m.Date = time.Now()
	m.Content = api

	err = m.Insert()
	if err != nil {
		logs.Error("failed to insert chat message %v", err)
		c.JsonResult(1, err.Error(), err)
		return
	}

	inferenceServerUrl, err := web.AppConfig.String("inference_server_host")
	if err != nil {
		logs.Error("failed to get inference server host %v", err)
		c.JsonResult(1, "获取推理服务地址失败", err)
		return
	}
	body := map[string]any{
		"data": doc.Markdown,
	}
	request := httplib.Post(inferenceServerUrl + api)
	request.JSONBody(body)
	request.Header("Content-Type", "application/json").Response()
	m.Response, err = request.String()
	if err != nil {
		logs.Error("failed to call inference server %v", err)
		c.JsonResult(1, "访问推理服务失败", err)
	}
	logs.Trace("inference result %s", m.Response)
	m.Update("response")

	c.JsonResult(0, "ok", m)
}

func (c *AigcController) DocChat() {
	prompt := c.GetString("prompt")
	id, _ := c.GetInt("doc_id")
	logs.Notice("prompt %s doc_id %d", prompt, id)

	doc, err := models.NewDocument().Find(id)
	if err != nil {
		c.JsonResult(1, "文章不存在")
		return
	}

	m := models.NewAigcChatMessage()
	m.DocumentId = id
	if !c.isUserLoggedIn() {
		c.JsonResult(1, "请先登录，再进行交互")
		return
	}
	if len(c.Member.RealName) != 0 {
		m.Author = c.Member.RealName
	} else {
		m.Author = c.Member.Account
	}
	m.MemberId = c.Member.MemberId
	m.IPAddress = c.Ctx.Request.RemoteAddr
	m.IPAddress = strings.Split(m.IPAddress, ":")[0]
	m.Date = time.Now()
	m.Content = prompt
	err = m.Insert()
	if err != nil {
		logs.Error("failed to insert chat message %v", err)
		c.JsonResult(1, err.Error(), err)
		return
	}

	inferenceServerUrl, err := web.AppConfig.String("inference_server_host")
	if err != nil {
		logs.Error("failed to get inference server host %v", err)
		c.JsonResult(1, "获取推理服务地址失败", err)
		return
	}
	body := map[string]any{
		"input":    doc.Markdown,
		"question": prompt,
	}
	request := httplib.Post(inferenceServerUrl + "/api/get_qa")
	request.JSONBody(body)
	request.Header("Content-Type", "application/json").Response()
	m.Response, err = request.String()
	if err != nil {
		logs.Error("failed to call inference server %v", err)
		c.JsonResult(1, "访问推理服务失败", err)
	}
	logs.Trace("inference result %s", m.Response)
	m.Update("response")

	c.JsonResult(0, "ok", m)
}

func (c *AigcController) Index() {
	c.Prepare()
	c.TplName = "aigc/index.html"
}

func (c *AigcController) DeleteMessage() {
	c.Prepare()
	id, _ := c.GetInt(":id", 0)
	m, err := models.NewAigcChatMessage().Find(id)
	if err != nil {
		c.JsonResult(1, "消息不存在")
	}

	doc, err := models.NewDocument().Find(m.DocumentId)
	if err != nil {
		c.JsonResult(1, "文章不存在")
	}

	// 判断是否有权限删除
	bookRole, _ := models.NewRelationship().FindForRoleId(doc.BookId, c.Member.MemberId)
	if m.CanDelete(c.Member.MemberId, bookRole) {
		err := m.Delete()
		if err != nil {
			c.JsonResult(1, "删除错误")
		} else {
			c.JsonResult(0, "ok")
		}
	} else {
		c.JsonResult(1, "没有权限删除")
	}
}
