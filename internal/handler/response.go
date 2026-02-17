package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// Success 200 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 200, Msg: "success", Data: data})
}

// Created 201 创建成功响应
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Code: 201, Msg: "success", Data: data})
}

// Page 分页列表响应
func Page(c *gin.Context, list interface{}, total, page, pageSize int) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Msg:  "success",
		Data: PageData{List: list, Total: total, Page: page, PageSize: pageSize},
	})
}

// Fail 错误响应
func Fail(c *gin.Context, code int, msg string) {
	c.JSON(code, Response{Code: code, Msg: msg, Data: nil})
}
