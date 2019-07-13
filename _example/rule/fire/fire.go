package firenews

import (
	"encoding/json"
	"fmt"
	"github.com/mybee/go-spider/spider"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func init() {
	spider.Register(rule)
}

var (
	outputFields  = []string{"category", "type", "question", "options", "help", "chose"}
	outputFields2 = []string{"bigcategory", "category", "category_link"}

	namespace1 = "fire_question"
	namespace2 = "fire_category"
)
var multNamespaceConf = map[string]*spider.MultipleNamespaceConf{
	namespace1: {
		OutputFields:      outputFields,
		OutputConstraints: spider.NewStringsConstraints(outputFields, 64, 512, 712, 512, 512, 60),
	},
	namespace2: {
		OutputFields:      outputFields2,
		OutputConstraints: spider.NewStringsConstraints(outputFields2, 300, 256, 40),
	},
}

// 演示如何在一条规则里面，同时需要导出数据到两张表
var rule = &spider.TaskRule{
	Name:                      "fire规则",
	Description:               "fire",
	OutputToMultipleNamespace: true,
	MultipleNamespaceConf:     multNamespaceConf,
	Rule: &spider.Rule{
		Head: func(ctx *spider.Context) error {
			return ctx.VisitForNext("https://h5.yixianjingcheng.com/examMobile/mobile/uc/exam/getKnowledgePoint")
		},
		Nodes: map[int]*spider.Node{
			0: step1, // 第一步: 获取所有分类
			//1: step2, // 第二步: 获取每个分类的新闻标题链接
			1: step3,
		},
	},
}

var step1 = &spider.Node{
	OnRequest: func(ctx *spider.Context, req *spider.Request) {
		req.Headers.Add("cookie", "JSESSIONID=33767EBDE1E33D1E6EB0EC3B2B12695F; sid=C3C24C6D51B0EC98E1F4544F37540C14; route=fa0de6009ef9bf245531a1f7e35666e6; JSESSIONID=EC83A3FC06B28D709F3D26803E2105FE")
		log.Println("Visting", req.URL.String())
	},
	OnHTML: map[string]func(*spider.Context, *spider.HTMLElement) error{
		`.loresetExam`: func(ctx *spider.Context, el *spider.HTMLElement) error {

			bigCate := ""
			el.ForEach(".wm-title-txt", func(i int, element *spider.HTMLElement) {
				bigCate = element.Text
			})

			el.ForEach(".item-inner .item-title span", func(i int, element *spider.HTMLElement) {
				category := element.Text
				ctx.PutReqContextValue("category", category)
				if category != "首页" {
					err := ctx.Output(map[int]interface{}{
						0: bigCate,
						1: category,
						2: "",
					}, namespace2)
					if err != nil {
						fmt.Println(err)
					}
				}
			})

			id := el.Attr("data-id")
			fmt.Println("🍋", id)
			req, err := http.NewRequest("POST", "https://h5.yixianjingcheng.com/examMobile/mobile/uc/exam/ajax/toLorePointExam?lorePoint="+id, nil)
			if err != nil {
				fmt.Println("🍒", err)
				return err
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("cookie", "JSESSIONID=33767EBDE1E33D1E6EB0EC3B2B12695F; sid=C3C24C6D51B0EC98E1F4544F37540C14; route=fa0de6009ef9bf245531a1f7e35666e6; JSESSIONID=EC83A3FC06B28D709F3D26803E2105FE")
			// /mobile/uc/exam/ajax/toLorePointExam
			client := http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("🍳", err)
				return err
			}
			fmt.Println("🧀", resp)
			type result struct {
				Entity int `json:"entity"`
			}
			var re result
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("🍔", err)
				return err
			}
			fmt.Println("🍇", string(body))
			err = json.Unmarshal(body, &re)
			if err != nil {
				fmt.Println("🍖", err)
				return err
			}
			fmt.Println("🍇", re)
			time.Sleep(2 * time.Second)
			return ctx.VisitForNextWithContext("https://h5.yixianjingcheng.com/examMobile/mobile/uc/exam?examRecordId=" + strconv.Itoa(re.Entity))
		},
	},
}

var step3 = &spider.Node{
	OnRequest: func(ctx *spider.Context, req *spider.Request) {
		req.Headers.Add("cookie", "JSESSIONID=801A182EC276F3CC46E28322C657E5B2; sid=C3C24C6D51B0EC98E1F4544F37540C14; route=fa0de6009ef9bf245531a1f7e35666e6; JSESSIONID=EC83A3FC06B28D709F3D26803E2105FE")
		log.Println("Visting", req.URL.String())
	},
	OnHTML: map[string]func(*spider.Context, *spider.HTMLElement) error{
		`.swiper-slide`: func(ctx *spider.Context, el *spider.HTMLElement) error {
			question := ""
			options := ""
			help := ""
			chose := ""
			qtype := el.Attr("data-type")

			isOptions := false

			el.ForEach(".question-info-txt__wrap div p", func(i int, element *spider.HTMLElement) {
				element.ForEach("span", func(i int, eleme *spider.HTMLElement) {
					question += eleme.Text
				})
			})

			el.ForEach(".opTxt span p", func(i int, element *spider.HTMLElement) {

				options += element.Text + "&&"
				isOptions = true
			})

			el.ForEach(".opTxt p span", func(i int, element *spider.HTMLElement) {
				if !isOptions {
					options += element.Text + "&&"
				}
			})

			el.ForEach(".c-666 p", func(i int, element *spider.HTMLElement) {
				help = element.Text
			})

			el.ForEach(".vam strong", func(i int, element *spider.HTMLElement) {
				chose = element.Text
			})

			category := ctx.GetReqContextValue("category")
			return ctx.Output(map[int]interface{}{
				0: category,
				1: qtype,
				2: question,
				3: options,
				4: help,
				5: chose,
			}, namespace1)
		},
	},
}
