package main

import (
	"fmt"
	"golang.org/x/net/html"
	"net/http"
	"os"
	"runtime"
	"strings"
)

// Тестовое задание "сделать парсер 1000-и и более страниц при условии 4 ядер"!

// Вспомогательная функция для извлечения атрибута href из токена
func getHref(t html.Token) (ok bool, href string) {
	// Перебираем атрибуты токена, пока не найдем "href"
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}

	return
}

// Извлечь все ссылки http ** с данной веб-страницы
func ulrParser(url string, ch chan string, chFinished chan bool) {
	resp, err := http.Get(url)

	defer func() {
		// Сообщите, что мы закончили после этой функции
		chFinished <- true
	}()

	if err != nil {
		fmt.Println("ERROR: Failed to ulrParser:", url)
		return
	}

	b := resp.Body
	defer b.Close() // закрыть тело, когда функция завершится

	z := html.NewTokenizer(b)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:

			// конец документа, мы закончили...

			return
		case tt == html.StartTagToken:
			t := z.Token()

			// проверка, является ли токен тегом <a>
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// извлекаем значение href, если оно есть
			ok, url := getHref(t)
			if !ok {
				continue
			}

			// проверка, что URL-адрес начинается с http **
			hasProto := strings.Index(url, "http") == 0
			if hasProto {
				ch <- url
			}
		}
	}
}

func main() {
	// Устанавливаем макс.кол-во используемых ядер в runtime для наиболее эффективного параллелизма
	runtime.GOMAXPROCS(4)
	foundUrls := make(map[string]bool)
	seedUrls := os.Args[1:]

	// каналы
	chUrls := make(chan string)
	chFinished := make(chan bool)

	// начало параллельного сканирования
	for _, url := range seedUrls {
		go ulrParser(url, chUrls, chFinished)
	}

	// для параллельного исполнения используем оператор select
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			foundUrls[url] = true
		case <-chFinished:
			c++
		}
	}

	// распечатываем результаты ...

	fmt.Println("\nFound", len(foundUrls), "unique urls:\n")

	for url, _ := range foundUrls {
		fmt.Println(" - " + url)
	}

	close(chUrls)
}