package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

type Task struct {
	ID          int
	Description string
}

func main() {
	var err error

	log.SetOutput(os.Stdout)

	db, err = gorm.Open(sqlite.Open("todo-list.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&Task{}); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("GET /", getRoot)
	http.HandleFunc("POST /add-task", postAddTask)
	http.HandleFunc("DELETE /delete-task/", deleteTask)

	log.Fatal(http.ListenAndServe("127.0.0.1:3000", nil))
}

func Render(wr io.Writer, name string, data any) error {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(wr, name, data)
}
func serveStaticContent(w http.ResponseWriter, r *http.Request) {
	filePath := path.Join("web", r.URL.Path)

	if _, err := os.Stat(filePath); err != nil {
		log.Println(err)
		code := http.StatusNotFound
		http.Error(w, http.StatusText(code), code)
		return
	}

	http.ServeFile(w, r, filePath)
}

func getRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		serveStaticContent(w, r)
		return
	}

	var tasks []*Task

	if err := db.Model(&Task{}).Find(&tasks).Error; err != nil {
		log.Println(err)
		code := http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
		return
	}

	if err := Render(w, "index.html", map[string]any{
		"Tasks": tasks,
	}); err != nil {
		log.Println(err)
		code := http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
		return
	}
}

func postAddTask(w http.ResponseWriter, r *http.Request) {
	userTask := r.PostFormValue("task")

	newTask := &Task{
		Description: userTask,
	}

	if err := db.Model(&Task{}).Create(&newTask).Error; err != nil {
		log.Println(err)
		code := http.StatusInternalServerError
		http.Error(w, http.StatusText(code), code)
		return
	}

	if err := Render(w, "app-task-item", newTask); err != nil {
		log.Println(err)
	}
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(r.URL.Path, "/")
	urlID := segments[len(segments)-1]
	id, err := strconv.Atoi(urlID)
	if err != nil {
		log.Println(err)
		code := http.StatusBadRequest
		http.Error(w, http.StatusText(code), code)
		return
	}

	tx := db.Delete(&Task{}, id)
	if tx.Error != nil && tx.RowsAffected == 0 {
		log.Println(err)
		code := http.StatusNotFound
		http.Error(w, http.StatusText(code), code)
		return
	}

	if _, err := w.Write([]byte(" ")); err != nil {
		log.Println(err)
	}
}
