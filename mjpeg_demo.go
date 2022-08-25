package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

//go:embed templates/*
var resources embed.FS

var t = template.Must(template.ParseFS(resources, "templates/*"))

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex)
	mux.HandleFunc("/count.mjpeg", handleMJPEG)

	log.Println("listening on", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "text/html; charset=utf-8")
	err := t.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Printf("Execute template err: %s", err)
	}
}

func handleMJPEG(w http.ResponseWriter, r *http.Request) {
	log.Printf("handle mjpeg")
	h := w.Header()
	h.Set("Content-Type", "multipart/x-mixed-replace; boundary=myboundary")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "close")
	h.Set("Pragma", "no-cache")

	ctx := r.Context()

	flusher := w.(http.Flusher)

OUTER:
	for i := 0; i < 100; i++ {
		log.Printf("genIMage %d", i)
		img, err := genImage(i)
		if err != nil {
			log.Fatalf("genImg err: %s", err)
		}

		fmt.Fprintf(w, "--myboundary\r\n")
		fmt.Fprintf(w, "Content-Type: image/jpeg\r\n")
		fmt.Fprintf(w, "Content-Length: %d\r\n", len(img))
		fmt.Fprintf(w, "\r\n")
		w.Write(img)
		fmt.Fprintf(w, "\r\n")
		flusher.Flush()

		select {
		case <-time.After(500 * time.Millisecond):
		case <-ctx.Done():
			break OUTER
		}
	}
}

func genImage(i int) ([]byte, error) {
	width := 100
	height := width

	bg := color.RGBA{0xad, 0xd9, 0xe6, 0xff}

	m := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(m, m.Bounds(), &image.Uniform{bg}, image.ZP, draw.Src)

	offset := width / 2
	addLabel(m, offset, offset, strconv.Itoa(i))

	// 0002ff
	var b bytes.Buffer
	err := jpeg.Encode(&b, m, nil)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{0x00, 0x02, 0xff, 255}
	point := fixed.Point26_6{fixed.I(x), fixed.I(y)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	d.DrawString(label)
}
