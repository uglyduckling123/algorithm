package controller

import "C"
import (
	"context"
	"encoding/json"
	"fmt"
	"mango/internal/service"
	"net/http"
	"strings"
	"time"

	"github.com/volcengine/volc-sdk-golang/service/businessSecurity"

	"github.com/gin-gonic/gin"
)

type VolcHandler struct {
	userService service.UserService
	textService service.TextRiskLogService
}

func init() {
	businessSecurity.DefaultInstance.Client.SetAccessKey("AKLTYTRhOGQ0OTY2NWYxNGI4NThmMDBiMTc5MjUyODE0MGU")
	businessSecurity.DefaultInstance.Client.SetSecretKey("TlRnMVpXSmxOVFJqWVdNM05EazVORGxtTXpKbVpEbGhaVEl3TTJSa04yRQ==")
}
func NewVolcHandler(userService service.UserService, textService service.TextRiskLogService) *VolcHandler {
	return &VolcHandler{
		userService: userService,
		textService: textService,
	}
}

// Register 注册路由
func (v *VolcHandler) Register(router *gin.RouterGroup) {
	userRouter := router.Group("/volc")
	{
		userRouter.POST("/text", v.text)
		userRouter.GET("/img", v.img)
	}
}

type SnRequest struct {
	Text string `json:"text" form:"text" binding:"required"`
}
type HuoShanTextCheckRequest struct {
	AccountID   string `json:"account_id"`
	BizType     string `json:"biztype"`
	Text        string `json:"text"`
	OperateTime int64  `json:"operate_time"`
	TextType    string `json:"text_type"`
}

func (v *VolcHandler) text(c *gin.Context) {
	var request SnRequest
	err := c.ShouldBind(&request)
	if err != nil {

	}

	params := HuoShanTextCheckRequest{
		AccountID:   "752431",
		BizType:     "aippt_cn",
		Text:        request.Text,
		OperateTime: time.Now().Unix(),
		TextType:    "prompt",
	}
	requestData, _ := json.Marshal(params)
	result, _ := businessSecurity.DefaultInstance.TextSliceRisk(&businessSecurity.RiskDetectionRequest{
		AppId:      752431,
		Service:    "text_risk",
		Parameters: string(requestData)})
	fmt.Println("=========================result")
	fmt.Println(result)
	c.JSON(http.StatusCreated, result)
}

const MaxRiskText = 6000

func (v *VolcHandler) TextChunkByWrap(ctx context.Context, text, wrap string, maxSize int) []string {
	var texts []string
	//小于最大的数量
	if len([]rune(text)) <= maxSize {
		return append(texts, text)
	}

	split := strings.Split(text, wrap)
	//无法分隔
	if len(split) == 1 {
		return append(texts, text)
	}
	tmp := ""
	for i, t := range split {
		//累计加起来大于max 推入texts  重置tmp
		if len([]rune(tmp+wrap+t)) > maxSize {
			texts = append(texts, tmp)
			tmp = ""
		}

		//单行数据拼接到tmp
		tmp = tmp + t + wrap
		//最后一次循环 最后一个tmp 推入texts
		if i+1 >= len(split) {
			texts = append(texts, tmp)
		}

	}
	return texts
}

type HuoShanImgCheckRequest struct {
	AccountID   string `json:"account_id"`
	BizType     string `json:"biztype"`
	Url         string `json:"url"`
	DataId      string `json:"data_id"`
	OperateTime int64  `json:"operate_time"`
	PictureType string `json:"picture_type"`
}

func (v *VolcHandler) img(c *gin.Context) {
	params := &HuoShanImgCheckRequest{
		AccountID:   "752431",
		BizType:     "aippt_cn",
		Url:         "https://aippt-domestic-test.aippt.com/365sheji-user/headImage/4b199cd69f0e467bb4837a72f4618152.jpeg",
		DataId:      "1130011747808932",
		OperateTime: time.Now().Unix(),
		PictureType: "prompt",
	}
	requestData, _ := json.Marshal(params)
	result, _ := businessSecurity.DefaultInstance.ImageContentRiskV2(&businessSecurity.RiskDetectionRequest{
		AppId:      752431,               // write your app id
		Service:    "image_content_risk", // write business security service
		Parameters: string(requestData),  // write your parameters
	})
	c.JSON(http.StatusCreated, result)
}
