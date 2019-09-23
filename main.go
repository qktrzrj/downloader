package main

func main() {
	url := "https://s3.amazonaws.com/SQLyog_Community/SQLyog+13.1.5/SQLyog-13.1.5-0.x64Community.exe"
	err := download(url)
	if err != nil {
		panic(err)
	}
}
