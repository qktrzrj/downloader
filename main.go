package main

func main() {
	url := "http://down.sandai.net/mac/thunder_3.3.7.4170.dmg"
	err := Download(url)
	if err != nil {
		panic(err)
	}
}
