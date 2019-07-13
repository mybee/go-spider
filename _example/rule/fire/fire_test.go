package firenews

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
)

type Point struct {
	LorePoint string `json:"lorePoint"`
}

func TestFire(t *testing.T) {

	req, err := http.NewRequest("POST", "https://h5.yixianjingcheng.com/examMobile/mobile/uc/exam/ajax/toLorePointExam?lorePoint=146", nil)
	if err != nil {
		fmt.Println("🍒", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("cookie", "JSESSIONID=33767EBDE1E33D1E6EB0EC3B2B12695F; sid=C3C24C6D51B0EC98E1F4544F37540C14; route=fa0de6009ef9bf245531a1f7e35666e6; JSESSIONID=EC83A3FC06B28D709F3D26803E2105FE")
	// /mobile/uc/exam/ajax/toLorePointExam
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("🍳", err)
	}
	fmt.Println("🧀", resp)
	type result struct {
		Entity  int  `json:"entity"`
		Success bool `json:"success"`
	}
	var re result
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("🍔", err)
	}
	fmt.Println("🍇", string(body))
	err = json.Unmarshal(body, &re)
	if err != nil {
		fmt.Println("🍖", err)
	}
	fmt.Println("🍇", re)
}
