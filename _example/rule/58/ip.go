package ip

import (
	"fmt"
	"github.com/mybee/go-spider/spider"
	log "github.com/sirupsen/logrus"
	"strings"
	"unicode"
)

func init() {
	spider.Register(rule)
}

// 演示如何在一条规则里面，同时需要导出数据到两张表
var rule = &spider.TaskRule{
	Name:                      "58city",
	Description:               "抓取58的访客ip",
	OutputToMultipleNamespace: true,
	MultipleNamespaceConf:     nil,
	Rule: &spider.Rule{
		Head: func(ctx *spider.Context) error {
			return ctx.VisitForNext("https://vistorip.58.com/vistors?r=0.6552303818508352&r=0.10731092311874646")
		},
		Nodes: map[int]*spider.Node{
			0: step1, // 第一步: 获取所有分类
		},
	},
}

//type

var step1 = &spider.Node{
	OnRequest: func(ctx *spider.Context, req *spider.Request) {
		req.Headers.Add("cookie", `commontopbar_new_city_info=1%7C%E5%8C%97%E4%BA%AC%7Cbj; 58home=bj; id58=e87rZl0oGrCHtJkuBu7xAg==; city=bj; 58tj_uuid=265138eb-f40b-4369-b118-1613e53f6d22; xxzl_deviceid=Zze7Ffak9p3YV4%2B%2B38bpAqN9E0Npd0vApWEkX19CAH5kUM83OochCvmRYdXPLH8w; PPU="UID=18297784695559&UN=sk1136&TT=83035cf98bf71733c5a89f01bf922434&PBODY=PJUGyNwWFQudjSehAB0iL_xEdj8HK6VqWYe4lDHT9HtjbVlMLJfjvEUVXjlpJXx4O6fRZ1-vhhOtzdWXOkYHcSAMSP0ZD8ZAW8MhHcY1DVlFLOhjQoyEBqywn0F12PCpOgJ_Zxm7vQebIhPLecPi2Ol-bOCmOMZljzBQ2IgoZSg&VER=1"; www58com="UserID=18297784695559&UserName=sk1136"; 58cooper="userid=18297784695559&username=sk1136"; 58uname=sk1136; vip=vipusertype%3D0%26vipuserpline%3D0%26v%3D1%26vipkey%3D4c68eaf085168fbfcc75f3b32ceb113c%26masteruserid%3D18297784695559; wmda_uuid=24ec84b257000f251e385b7210dddec3; wmda_new_uuid=1; wmda_session_id_2286118353409=1562914706990-616d6d04-216a-adae; wmda_visited_projects=%3B2286118353409; new_session=1; new_uv=2; utm_source=; spm=; init_refer=https%253A%252F%252Fpassport.58.com%252Fsec%252F58%252Ffeature%252Fpc%252Fui%253Fwarnkey%253DffVFc4iaKI6aQHXLGHo4mNwgCoAdplP4%2526path%253Dhttps%25253A%25252F%25252Fvip.58.com%25252Fvcenter%25252Fvisitor%25252F%25253Fr%25253D0.39223988041820257%252526pts%25253D1562914666507%2526source%253D58-default-pc%2526requesthost%253Dpassport.58.com%2526domain%253D58.com; als=0; xxzl_smartid=83fd2d359c2391cdd591cc57843f8877`)
		log.Println("Visting", req.URL.String())
	},
	OnHTML: map[string]func(*spider.Context, *spider.HTMLElement) error{
		`body script`: func(ctx *spider.Context, el *spider.HTMLElement) error { // 获取所有分类

			fmt.Println(el.Text)
			text := strings.FieldsFunc(el.Text, unicode.IsSpace)
			fmt.Println(text)
			return nil
		},
	},
}
