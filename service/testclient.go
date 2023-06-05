package service

import "net/http"

//測試打第三方API
func TestThirdAPI() {
	req, _ := http.NewRequest("GET", "url", nil)
	req.Header.Set("", "")
	http.DefaultClient.Do(req)
}

//測試併發
func TestMulti() {
	for i := 1; i <= 1000; i++ {
		go TestThirdAPI()
	}
}
