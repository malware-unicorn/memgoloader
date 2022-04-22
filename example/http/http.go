package http

import (
	"log"
    "io/ioutil"
    "net/http"
    "fmt"
    "bytes"
)

func HandleSend(domain string, msg []byte) []byte {
    symbolurl := fmt.Sprintf("http://%s", domain)
    log.Printf("\033[33mTR: HandleSend() %s\033[0m", symbolurl)
    client := http.Client{}
    resp, err := client.Post(symbolurl, "image/jpeg", bytes.NewBuffer(msg))
    if err != nil{
        log.Printf("\033[33mTR: Failed Request %s\033[0m", err.Error())
        return nil
    }
    if resp.StatusCode != 200 {
        log.Printf("\033[33mTR: Failed Request \033[0m")
        return nil
    }
    log.Printf("\033[33mTR: HandleSend() Afterpost\033[0m")

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
    	log.Printf("\033[33mTR: Failed Read\033[0m")
        resp.Body.Close()
        return nil
    }
    //resp.Body.Close()
    log.Printf("\033[33mTR: body len %v\033[0m", body)
    return body

    return []byte{}
}

func main() {
    log.Printf("Hello")
    HandleSend("localhost:8080", []byte{})

}
