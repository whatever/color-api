package main

import (
	"context"
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

// Run the trap
func SetupTrap(server *http.Server) {
	cigs := make(chan os.Signal, 1)
	signal.Notify(cigs, syscall.SIGINT)
	go func() {
		<-cigs
		if err := server.Shutdown(context.Background()); err != nil {
			fmt.Println(err)
		}
	}()
}

type Color [3]uint8

func (c Color) String() string {
	return fmt.Sprintf("#%02x%02x%02x", c[0], c[1], c[2])
}

func (c Color) Xyz() string {
	return fmt.Sprintf("%02x%02x%02x", c[0], c[1], c[2])
}

func RandomColor(rand *rand.Rand) Color {
	return Color{uint8(rand.Intn(256)), uint8(rand.Intn(256)), uint8(rand.Intn(256))}
}

func ColorFromString(s string) Color {
	if len(s) != 6 {
		return Color{255, 255, 255}
	}

	r, _ := strconv.ParseInt(s[0:2], 16, 64)
	g, _ := strconv.ParseInt(s[2:4], 16, 64)
	b, _ := strconv.ParseInt(s[4:6], 16, 64)

	return Color{uint8(r), uint8(g), uint8(b)}
}

func (c *Color) Image() image.Image {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	draw.Draw(
		img,
		img.Bounds(),
		&image.Uniform{color.RGBA{c[0], c[1], c[2], 255}},
		image.ZP,
		draw.Src,
	)
	return img
}

func main() {

	port := flag.Int("port", 8080, "port number to open on")
	html := flag.String("html", "index.html", "location of file to use as a template")
	flag.Parse()

	sour := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(sour)

	s := &http.Server{
		Addr: fmt.Sprintf(":%v", *port),
	}

	fname := *html

	tmpl, err := template.New("color").ParseFiles(fname)

	if err != nil {
		fmt.Printf("error: %v\n", err)
		return
	}

	SetupTrap(s)

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c := RandomColor(rand)
		fmt.Println("Random color:", c)
		if err := tmpl.ExecuteTemplate(w, "index.html", c); err != nil {
			fmt.Println(err)
		}
	})

	r.HandleFunc("favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Vanilla favicon.ico...")
	})

	r.HandleFunc("/{color}/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		c := ColorFromString(mux.Vars(r)["color"])
		png.Encode(w, c.Image())
	})

	r.HandleFunc("/color", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%v", RandomColor(rand))
	})

	http.Handle("/", r)

	fmt.Println("Finished:", s.ListenAndServe())
}
